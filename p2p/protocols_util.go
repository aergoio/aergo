/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import "strings"

var subProtocolMap map[string]SubProtocol
var subProtocolCodeMap map[uint32]SubProtocol

func init() {
	subProtocolMap = make(map[string]SubProtocol)
	subProtocolCodeMap = make(map[uint32]SubProtocol)
	for i:=StatusRequest; i<=0x030 ; i++ {
		if strings.HasPrefix(i.String(),"SubProtocol(") {
			continue
		}
		subProtocolMap[i.String()] = i
		subProtocolCodeMap[i.Uint32()] = i
	}
}

func ParseSubProtocol(str string) (SubProtocol, bool) {
	sp, found := subProtocolMap[str]
	return sp, found
}

func FindSubProtocol(code uint32) (SubProtocol, bool) {
	sp, found := subProtocolCodeMap[code]
	return sp, found
}
