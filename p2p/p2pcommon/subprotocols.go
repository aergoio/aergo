/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// NOTE: change const of protocols_test.go
const (
	_ SubProtocol = 0x00 + iota
	StatusRequest
	PingRequest
	PingResponse
	GoAway
	AddressesRequest
	AddressesResponse
)
const (
	GetBlocksRequest SubProtocol = 0x010 + iota
	GetBlocksResponse
	GetBlockHeadersRequest
	GetBlockHeadersResponse
	_ // placeholder for deprecated GetMissingRequest
	_ // placeholder for deprecated GetMissingResponse
	NewBlockNotice
	GetAncestorRequest
	GetAncestorResponse
	GetHashesRequest
	GetHashesResponse
	GetHashByNoRequest
	GetHashByNoResponse
)
const (
	GetTXsRequest SubProtocol = 0x020 + iota
	GetTXsResponse
	NewTxNotice
)

// subprotocols for block producers and their own trusted nodes
const (
	// BlockProducedNotice from block producer to trusted nodes and other bp nodes
	BlockProducedNotice SubProtocol = 0x030 + iota
)

const (
	_ SubProtocol = 0x3100 + iota
	GetClusterRequest
	GetClusterResponse
	RaftWrapperMessage  //
)

//go:generate stringer -type=SubProtocol
