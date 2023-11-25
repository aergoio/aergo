/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"time"

	"github.com/aergoio/aergo/v2/types"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
)

// NTContainer can provide NetworkTransport interface.
//
//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/p2p/p2pcommon NTContainer,NetworkTransport | sed -e 's/^package mock_p2pcommon/package p2pmock/g' > ../p2pmock/mock_networktransport.go"
type NTContainer interface {
	GetNetworkTransport() NetworkTransport

	// GenesisChainID return inititial chainID of current chain.
	GenesisChainID() *types.ChainID
	SelfMeta() PeerMeta
}

// NetworkTransport do manager network connection
type NetworkTransport interface {
	core.Host
	Start() error
	Stop() error

	SelfMeta() PeerMeta

	GetAddressesOfPeer(peerID types.PeerID) []string

	// AddStreamHandler wrapper function which call host.SetStreamHandler after transport is initialized, this method is for preventing nil error.
	AddStreamHandler(pid core.ProtocolID, handler network.StreamHandler)

	GetOrCreateStream(meta PeerMeta, protocolIDs ...core.ProtocolID) (core.Stream, error)
	GetOrCreateStreamWithTTL(meta PeerMeta, ttl time.Duration, protocolIDs ...core.ProtocolID) (core.Stream, error)

	FindPeer(peerID types.PeerID) bool
	ClosePeerConnection(peerID types.PeerID) bool
}
