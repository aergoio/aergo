/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

// constants of p2p protocol since v0.3
const (
	// this magic number is useful only in handshaking
	MAGICMain uint32 = 0x47416841
	MAGICTest uint32 = 0x2e415429

	P2PVersion030 uint32 = 0x00000300

	SigLength = 16
	IDLength = 16

	MaxPayloadLength = 1 << 23  // 8MB

	MaxBlockHeaderFetchSize = 10000
)


// SubProtocol identifies the type of p2p message
type SubProtocol uint32

//
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
	GetMissingRequest
	GetMissingResponse
	NewBlockNotice
)
const (
	GetTXsRequest SubProtocol = 0x020 + iota
	GetTxsResponse
	NewTxNotice
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
