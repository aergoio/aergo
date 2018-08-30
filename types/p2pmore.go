/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"net"
	"strconv"

	peer "github.com/libp2p/go-libp2p-peer"
)

// GetMessageData is delegation method for backward compatability
func (m *P2PMessage) GetMessageData() *MessageData {
	return m.Header
}

// AddressesToStringMap make map of string for logging or json encoding
func AddressesToStringMap(addrs []*PeerAddress) []map[string]string {
	arr := make([]map[string]string, len(addrs))
	for i, addr := range addrs {
		vMap := make(map[string]string)
		vMap["address"] = net.IP(addr.Address).String()
		vMap["port"] = strconv.Itoa(int(addr.Port))
		vMap["peerId"] = peer.ID(addr.PeerID).Pretty()
		arr[i] = vMap
	}
	return arr
}
