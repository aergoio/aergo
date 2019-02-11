/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

// SubProtocol identifies the lower type of p2p message
type SubProtocol uint32

func (i SubProtocol) Uint32() uint32 {
	return uint32(i)
}


