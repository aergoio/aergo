package subproto

import "github.com/aergoio/aergo/p2p/p2pcommon"

// NOTE: change const of protocols_test.go
const (
	_ p2pcommon.SubProtocol = 0x00 + iota
	StatusRequest
	PingRequest
	PingResponse
	GoAway
	AddressesRequest
	AddressesResponse
)
const (
	GetBlocksRequest p2pcommon.SubProtocol = 0x010 + iota
	GetBlocksResponse
	GetBlockHeadersRequest
	GetBlockHeadersResponse
	GetMissingRequest  // Deprecated
	GetMissingResponse // Deprecated
	NewBlockNotice
	GetAncestorRequest
	GetAncestorResponse
	GetHashesRequest
	GetHashesResponse
	GetHashByNoRequest
	GetHashByNoResponse
)
const (
	GetTXsRequest p2pcommon.SubProtocol = 0x020 + iota
	GetTXsResponse
	NewTxNotice
)

// subprotocols for block producers and their own trusted nodes
const (
	// BlockProducedNotice from block producer to trusted nodes and other bp nodes
	BlockProducedNotice p2pcommon.SubProtocol = 0x030 + iota
)

const (
	_ p2pcommon.SubProtocol = 0x3100 + iota
	GetClusterRequest
	GetClusterResponse
)

//go:generate stringer -type=SubProtocol
