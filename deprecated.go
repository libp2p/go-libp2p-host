package host

import (
	"github.com/libp2p/go-libp2p/helpers"

	moved "github.com/libp2p/go-libp2p/skel/host"
)

// Deprecated: use github.com/libp2p/go-libp2p/skel/host.Host instead.
type Host = moved.Host

// Deprecated: github.com/libp2p/go-libp2p/peer.InfoFromHost.
var PeerInfoFromHost = moved.InfoFromHost

// Deprecated: use helpers.MultistreamSemverMatcher.
var MultistreamSemverMatcher = helpers.MultistreamSemverMatcher