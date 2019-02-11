/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"strings"
)

var subProtocolMap map[string]p2pcommon.SubProtocol
var subProtocolCodeMap map[uint32]p2pcommon.SubProtocol

func init() {
	subProtocolMap = make(map[string]p2pcommon.SubProtocol)
	subProtocolCodeMap = make(map[uint32]p2pcommon.SubProtocol)
	for i:=StatusRequest; i<=0x030 ; i++ {
		if strings.HasPrefix(i.String(),"SubProtocol(") {
			continue
		}
		subProtocolMap[i.String()] = i
		subProtocolCodeMap[i.Uint32()] = i
	}
}

func ParseSubProtocol(str string) (p2pcommon.SubProtocol, bool) {
	sp, found := subProtocolMap[str]
	return sp, found
}

func FindSubProtocol(code uint32) (p2pcommon.SubProtocol, bool) {
	sp, found := subProtocolCodeMap[code]
	return sp, found
}
