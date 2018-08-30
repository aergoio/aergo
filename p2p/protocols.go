/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

// SubProtocol indentify the type of p2p message
type SubProtocol uint32

//
const (
	_ SubProtocol = 0x00 + iota
	statusRequest
	pingRequest
	pingResponse
	goAway
	addressesRequest
	addressesResponse
)
const (
	getBlocksRequest SubProtocol = 0x010 + iota
	getBlocksResponse
	getBlockHeadersRequest
	getBlockHeadersResponse
	getMissingRequest
	getMissingResponse
	newBlockNotice
)
const (
	getTXsRequest SubProtocol = 0x020 + iota
	getTxsResponse
	newTxNotice
)

//go:generate stringer -type=SubProtocol
