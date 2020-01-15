package contract

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"

	"golang.org/x/crypto/sha3"
)

const (
	shortNode  = 2
	branchNode = 17
	hexChar    = "0123456789abcdef"
)

type (
	rlpNode   [][]byte
	keyStream struct {
		*bytes.Buffer
	}
)

var (
	errDecode = errors.New("storage proof decode error")
	lenBuf    = make([]byte, 8)
	nilBuf    = make([]byte, 8)
)

func verifyEthStorageProof(key []byte, value rlpObject, expectedHash []byte, proof [][]byte) bool {
	if len(key) == 0 || value == nil || len(proof) == 0 {
		return false
	}
	key = []byte(hex.EncodeToString(keccak256(key)))
	valueRlpEncoded := rlpEncode(value)
	ks := keyStream{bytes.NewBuffer(key)}
	for i, p := range proof {
		if ((i != 0 && len(p) < 32) || !bytes.Equal(expectedHash, keccak256(p))) && !bytes.Equal(expectedHash, p) {
			return false
		}
		n := decodeRlpTrieNode(p)
		switch len(n) {
		case shortNode:
			if len(n[0]) == 0 {
				return false
			}
			leaf, sharedNibbles, err := decodeHpHeader(n[0][0])
			if err != nil {
				return false
			}
			sharedNibbles = append(sharedNibbles, []byte(hex.EncodeToString(n[0][1:]))...)
			if len(sharedNibbles) == 0 {
				return false
			}
			if leaf {
				return bytes.Equal(sharedNibbles, ks.key(-1)) && bytes.Equal(n[1], valueRlpEncoded)
			}
			if !bytes.Equal(sharedNibbles, ks.key(len(sharedNibbles))) {
				return false
			}
			expectedHash = n[1]
		case branchNode:
			if ks.Len() == 0 {
				return bytes.Equal(n[16], valueRlpEncoded)
			}
			k := ks.index()
			if k > 0x0f {
				return false
			}
			expectedHash = n[k]
		default:
			return false
		}
	}
	return false
}

func decodeRlpTrieNode(data []byte) rlpNode {
	var (
		dataLen = uint64(len(data))
		node    rlpNode
	)
	if dataLen == uint64(0) {
		return nil
	}
	switch {
	case data[0] >= 0xf8:
		lenLen := int(data[0]) - 0xf7
		l, err := decodeLen(data[1:], lenLen)
		if err != nil {
			return nil
		}
		if dataLen != uint64(1)+uint64(lenLen)+l {
			return nil
		}
		node = toList(data[1+lenLen:], l)
	case data[0] >= 0xc0:
		l := uint64(data[0]) - 0xc0
		if dataLen != uint64(1+l) {
			return nil
		}
		node = toList(data[1:], l)
	}
	return node
}

func decodeLen(data []byte, lenLen int) (uint64, error) {
	if len(data) <= lenLen || lenLen > 8 {
		return 0, errDecode
	}
	switch lenLen {
	case 1:
		return uint64(data[0]), nil
	default:
		start := int(8 - lenLen)
		copy(lenBuf[:], nilBuf[:start])
		copy(lenBuf[start:], data[:lenLen])
		return binary.BigEndian.Uint64(lenBuf), nil
	}
}

func toList(data []byte, dataLen uint64) rlpNode {
	var (
		node   rlpNode
		offset = uint64(0)
	)
	for {
		e, l, err := toString(data[offset:])
		if err != nil {
			return nil
		}
		node = append(node, e)
		offset += l
		if dataLen == offset {
			break
		}
		if dataLen < offset {
			return nil
		}
	}
	nodeLen := uint64(len(node))
	if nodeLen != uint64(2) && nodeLen != uint64(17) {
		return nil
	}
	return node
}

func toString(data []byte) ([]byte, uint64, error) {
	if len(data) == 0 {
		return nil, 0, errDecode
	}
	switch {
	case data[0] <= 0x7f: // a single byte
		return data[0:1], 1, nil
	case data[0] <= 0xb7: // string <= 55
		end := 1 + data[0] - 0x80
		return data[1:end], uint64(end), nil
	case data[0] <= 0xbf: // string > 55
		lenLen := data[0] - 0xb7
		l, err := decodeLen(data[1:], int(lenLen))
		if err != nil {
			return nil, 0, err
		}
		start := 1 + lenLen
		end := uint64(start) + l
		return data[start:end], end, nil
	default:
		return nil, 0, errDecode
	}
}

func keccak256(data ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}

func keccak256Hex(data ...[]byte) string {
	return hex.EncodeToString(keccak256(data...))
}

func decodeHpHeader(b byte) (bool, []byte, error) {
	switch b >> 4 {
	case 0:
		return false, []byte{}, nil
	case 1:
		return false, []byte{hexChar[b&0x0f]}, nil
	case 2:
		return true, []byte{}, nil
	case 3:
		return true, []byte{hexChar[b&0x0f]}, nil
	default:
		return false, []byte{}, errDecode
	}
}

func hexToIndex(c byte) (byte, error) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', nil
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, nil
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, errDecode
}

func (ks keyStream) index() byte {
	b, err := ks.ReadByte()
	if err != nil {
		return 0x10
	}
	i, err := hexToIndex(b)
	if err != nil {
		return 0x10
	}
	return i
}

func (ks keyStream) key(l int) []byte {
	if l == -1 {
		return ks.Buffer.Bytes()
	}
	return ks.Buffer.Next(l)
}

func rlpEncode(o rlpObject) []byte {
	return o.rlpEncode()
}

type rlpObject interface {
	rlpEncode() []byte
}

type rlpString []byte

func (s rlpString) rlpEncode() []byte {
	var rlpBytes []byte
	l := len(s)
	if l == 1 && s[0] < 0x80 {
		rlpBytes = append(rlpBytes, s[0])
	} else {
		rlpBytes = append(rlpBytes, rlpLength(l, 0x80)...)
		rlpBytes = append(rlpBytes, s...)
	}
	return rlpBytes
}

type rlpList []rlpObject

func (l rlpList) rlpEncode() []byte {
	var rlpBytes []byte
	for _, item := range l {
		rlpBytes = append(rlpBytes, item.rlpEncode()...)
	}
	length := rlpLength(len(rlpBytes), 0xc0)
	return append(length, rlpBytes...)
}

func rlpLength(dataLen int, offset byte) []byte {
	if dataLen < 56 {
		return []byte{byte(dataLen) + offset}
	} else if dataLen < math.MaxInt32 {
		var output []byte
		b := toBinary(dataLen)
		output = append(output, byte(len(b)+int(offset)+55))
		return append(output, b...)
	} else {
		return []byte{}
	}
}

func toBinary(d int) []byte {
	var b []byte
	for d > 0 {
		b = append([]byte{byte(d % 256)}, b...)
		d /= 256
	}
	return b
}
