package host

import (
	"github.com/libp2p/go-libp2p-core/helpers"

	moved "github.com/libp2p/go-libp2p-core/host"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/host.Host instead.
type Host = moved.Host

// Deprecated: github.com/libp2p/go-libp2p-core/peer.InfoFromHost.
var PeerInfoFromHost = moved.InfoFromHost

// Deprecated: use helpers.MultistreamSemverMatcher.
var MultistreamSemverMatcher = helpers.MultistreamSemverMatcher
