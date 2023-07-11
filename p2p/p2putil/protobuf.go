/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/golang/protobuf/proto"
)

func CalculateFieldDescSize(varSize int) int {
	switch {
	case varSize == 0:
		return 0
	case varSize < 128:
		return 2
	case varSize < 16384:
		return 3
	case varSize < 2097152:
		return 4
	case varSize < 268435456:
		return 5
	case varSize < 34359738368:
		return 6
	default:
		return 7
	}
}

func MarshalMessageBody(message p2pcommon.MessageBody) ([]byte, error) {
	return proto.Marshal(message)
}

func UnmarshalMessageBody(data []byte, msgData p2pcommon.MessageBody) error {
	return proto.Unmarshal(data, msgData)
}

func UnmarshalAndReturn(data []byte, msgData p2pcommon.MessageBody) (p2pcommon.MessageBody, error) {
	return msgData, proto.Unmarshal(data, msgData)
}
