/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/libp2p/go-libp2p-protocol"
)


// constants of p2p protocol since v0.3
const (
	// this magic number is useful only in handshaking
	MAGICMain uint32 = 0x47416841
	MAGICTest uint32 = 0x2e415429

	P2PVersion030 uint32 = 0x00000300

	SigLength = 16
	IDLength = 16

	MaxPayloadLength = 1 << 23  // 8MB

	MaxBlockHeaderResponseCount = 10000
	MaxBlockResponseCount       = 2000
)

// context of multiaddr, as higher type of p2p message
const (

	aergoP2PSub protocol.ID = "/aergop2p/0.2"

)


// SubProtocol identifies the lower type of p2p message
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
	GetAncestorRequest
	GetAncestorResponse
	GetHashesRequest
	GetHashesResponse
	GetHashByNoRequest
	GetHashByNoResponse
)
const (
	GetTXsRequest SubProtocol = 0x020 + iota
	GetTxsResponse
	NewTxNotice
)

//go:generate stringer -type=SubProtocol

func (i SubProtocol) Uint32() uint32 {
	return uint32(i)
}


const (
	txhashLen  = 32
	blkhashLen = 32

)

type BlkHash [blkhashLen]byte

func ParseToBlkHash(bSlice []byte) (BlkHash, error) {
	var hash BlkHash
	if len(bSlice) != blkhashLen {
		return hash, fmt.Errorf("parse error: invalid length")
	}
	copy(hash[:], bSlice)
	return hash, nil
}

func MustParseBlkHash(bSlice []byte) BlkHash {
	hash, err := ParseToBlkHash(bSlice)
	if err != nil {
		panic(err)
	}
	return hash
}

func (h BlkHash) String() string {
	return enc.ToString(h[:])
}

func (h BlkHash) Slice() []byte {
	return h[:]
}

type TxHash [txhashLen]byte

func ParseToTxHash(bSlice []byte) (TxHash, error) {
	var hash TxHash
	if len(bSlice) != txhashLen {
		return hash, fmt.Errorf("parse error: invalid length")
	}
	copy(hash[:], bSlice)
	return hash, nil
}

func MustParseTxHash(bSlice []byte) TxHash {
	hash, err := ParseToTxHash(bSlice)
	if err != nil {
		panic(err)
	}
	return hash
}

func (h TxHash) String() string {
	return enc.ToString(h[:])
}

func (h TxHash) Slice() []byte {
	return h[:]
}
