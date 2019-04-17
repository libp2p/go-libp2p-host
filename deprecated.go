package host

import (
	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	moved "github.com/libp2p/go-libp2p-core/host"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/host.Host instead.
type Host = moved.Host

// Deprecated: github.com/libp2p/go-libp2p-core/peer.InfoFromHost.
func PeerInfoFromHost(h Host) *peer.Info {
	return moved.InfoFromHost(h)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/helpers.MultistreamSemverMatcher.
func MultistreamSemverMatcher(base protocol.ID) (func(string) bool, error) {
	return helpers.MultistreamSemverMatcher(base)
}
