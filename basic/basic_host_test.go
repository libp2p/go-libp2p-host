package basichost_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	host "github.com/libp2p/go-libp2p/p2p/host"
	testutil "github.com/libp2p/go-libp2p/p2p/test/util"
)

func TestHostSimple(t *testing.T) {

	ctx := context.Background()
	h1 := testutil.GenHostSwarm(t, ctx)
	h2 := testutil.GenHostSwarm(t, ctx)
	defer h1.Close()
	defer h2.Close()

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal(err)
	}

	piper, pipew := io.Pipe()
	h2.SetStreamHandler(protocol.TestingID, func(s inet.Stream) {
		defer s.Close()
		w := io.MultiWriter(s, pipew)
		io.Copy(w, s) // mirror everything
	})

	s, err := h1.NewStream(ctx, h2pi.ID, protocol.TestingID)
	if err != nil {
		t.Fatal(err)
	}

	// write to the stream
	buf1 := []byte("abcdefghijkl")
	if _, err := s.Write(buf1); err != nil {
		t.Fatal(err)
	}

	// get it from the stream (echoed)
	buf2 := make([]byte, len(buf1))
	if _, err := io.ReadFull(s, buf2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf1, buf2) {
		t.Fatal("buf1 != buf2 -- %x != %x", buf1, buf2)
	}

	// get it from the pipe (tee)
	buf3 := make([]byte, len(buf1))
	if _, err := io.ReadFull(piper, buf3); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf1, buf3) {
		t.Fatal("buf1 != buf3 -- %x != %x", buf1, buf3)
	}
}

func getHostPair(ctx context.Context, t *testing.T) (host.Host, host.Host) {
	h1 := testutil.GenHostSwarm(t, ctx)
	h2 := testutil.GenHostSwarm(t, ctx)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal(err)
	}

	return h1, h2
}

func assertWait(t *testing.T, c chan protocol.ID, exp protocol.ID) {
	select {
	case proto := <-c:
		if proto != exp {
			t.Fatal("should have connected on ", exp)
		}
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for stream")
	}
}

func TestHostProtoPreference(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	protoOld := protocol.ID("/testing")
	protoNew := protocol.ID("/testing/1.1.0")
	protoMinor := protocol.ID("/testing/1.2.0")

	connectedOn := make(chan protocol.ID, 16)

	handler := func(s inet.Stream) {
		connectedOn <- s.Protocol()
		s.Close()
	}

	h1.SetStreamHandler(protoOld, handler)

	s, err := h2.NewStream(ctx, h1.ID(), protoMinor, protoNew, protoOld)
	if err != nil {
		t.Fatal(err)
	}

	assertWait(t, connectedOn, protoOld)
	s.Close()

	mfunc, err := host.MultistreamSemverMatcher(protoMinor)
	if err != nil {
		t.Fatal(err)
	}

	h1.SetStreamHandlerMatch(protoMinor, mfunc, handler)

	// remembered preference will be chosen first, even when the other side newly supports it
	s2, err := h2.NewStream(ctx, h1.ID(), protoMinor, protoNew, protoOld)
	if err != nil {
		t.Fatal(err)
	}

	// required to force 'lazy' handshake
	_, err = s2.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	assertWait(t, connectedOn, protoOld)

	s2.Close()

	s3, err := h2.NewStream(ctx, h1.ID(), protoMinor)
	if err != nil {
		t.Fatal(err)
	}

	assertWait(t, connectedOn, protoMinor)
	s3.Close()
}

func TestHostProtoMismatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	h1.SetStreamHandler("/super", func(s inet.Stream) {
		t.Error("shouldnt get here")
		s.Close()
	})

	_, err := h2.NewStream(ctx, h1.ID(), "/foo", "/bar", "/baz/1.0.0")
	if err == nil {
		t.Fatal("expected new stream to fail")
	}
}

func TestHostProtoPreknowledge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1 := testutil.GenHostSwarm(t, ctx)
	h2 := testutil.GenHostSwarm(t, ctx)

	conn := make(chan protocol.ID, 16)
	handler := func(s inet.Stream) {
		conn <- s.Protocol()
		s.Close()
	}

	h1.SetStreamHandler("/super", handler)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal(err)
	}
	defer h1.Close()
	defer h2.Close()

	// wait for identify handshake to finish completely
	time.Sleep(time.Millisecond * 20)

	h1.SetStreamHandler("/foo", handler)

	s, err := h2.NewStream(ctx, h1.ID(), "/foo", "/bar", "/super")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case p := <-conn:
		t.Fatal("shouldnt have gotten connection yet, we should have a lazy stream: ", p)
	case <-time.After(time.Millisecond * 50):
	}

	_, err = s.Read(nil)
	if err != nil {
		t.Fatal(err)
	}

	assertWait(t, conn, "/super")

	s.Close()
}

func TestNewDialOld(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h1, h2 := getHostPair(ctx, t)
	defer h1.Close()
	defer h2.Close()

	connectedOn := make(chan protocol.ID, 16)
	h1.SetStreamHandler("/testing", func(s inet.Stream) {
		connectedOn <- s.Protocol()
		s.Close()
	})

	s, err := h2.NewStream(ctx, h1.ID(), "/testing/1.0.0", "/testing")
	if err != nil {
		t.Fatal(err)
	}

	assertWait(t, connectedOn, "/testing")

	if s.Protocol() != "/testing" {
		t.Fatal("shoould have gotten /testing")
	}

	s.Close()
}
