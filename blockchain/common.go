/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"encoding/base64"
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

const (
	// DefaultMaxBlockSize is the maximum block size (currently 1MiB)
	DefaultMaxBlockSize = 1 << 20
)

func ToJSON(pb proto.Message) string {
	buf, _ := json.Marshal(pb)
	return string(buf)
}

// func ItobU64(argv uint64) []byte {
// 	bs := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(bs, argv)
// 	return bs
// }
// func BtoiU64(argv []byte) uint64 {
// 	is := binary.LittleEndian.Uint64(argv)
// 	return is
// }

func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

// func DecodeB64(sb string) []byte {
// 	buf, _ := base64.StdEncoding.DecodeString(sb)
// 	return buf
// }
