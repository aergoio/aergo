package types

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/base64"
)

const MAXBLOCKNO BlockNo = 18446744073709551615
const maxMetaSizeLimit = uint32(256 << 10)
const blockSizeHardLimit = uint32(8 << (10 * 2))

func EncodeB64(bs []byte) string {
	return base64.Encode(bs)
}

func DecodeB64(sb string) []byte {
	buf, _ := base64.Decode(sb)
	return buf
}

func EncodeB58(bs []byte) string {
	return base58.Encode(bs)
}

func DecodeB58(sb string) []byte {
	buf, _ := base58.Decode(sb)
	return buf
}

// GetMaxMessageSize returns the max message size corresponding to a specific block size (blockSize).
func GetMaxMessageSize(blockSize uint32) uint32 {
	return maxMetaSizeLimit + blockSize
}

// MaxMessageSize returns the limit for network message (client-server, p2p) size
func MaxMessageSize() uint32 {
	return GetMaxMessageSize(blockSizeHardLimit)
}

// BlockSizeHardLimit returns the hard limit for block size
func BlockSizeHardLimit() uint32 {
	return blockSizeHardLimit
}
