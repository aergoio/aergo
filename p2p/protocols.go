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

func (sp SubProtocol) Uint32() uint32 {
	return uint32(sp)
}

type SubProtocolMeta struct {
	SubProtocol
	flags spFlag
}

// flags of message
type spFlag int32

const (
	// spFlagExpectResponse is nearly same as request message. remote peer may store msgId of this type and can track if coresponing response is come
	spFlagExpectResponse spFlag = 1 << iota
	// spFlagResponse is this message is response, and has same msgID as coresponing request
	spFlagResponse
	// notypeSigned this message is signed.
	notypeSigned
	// spFlagSkippable means that message may skipped if receiver peer is busy. Most of newBlock or newTx notice is that, since next notice can drive peer in sync.
	spFlagSkippable
)
