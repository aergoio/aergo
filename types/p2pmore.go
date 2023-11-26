/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"encoding/json"
	"net"
	"strconv"
	"time"
)

type PeerBlockInfo interface {
	ID() PeerID
	State() PeerState
	LastStatus() *LastBlockStatus
}

// LastBlockStatus i
type LastBlockStatus struct {
	CheckTime   time.Time
	BlockHash   []byte
	BlockNumber uint64
}

// ResponseMessage contains response status
type ResponseMessage interface {
	GetStatus() ResultStatus
}

// AddressesToStringMap make map of string for logging or json encoding
func AddressesToStringMap(addrs []*PeerAddress) []map[string]string {
	arr := make([]map[string]string, len(addrs))
	for i, addr := range addrs {
		vMap := make(map[string]string)
		vMap["address"] = net.IP(addr.Address).String()
		vMap["port"] = strconv.Itoa(int(addr.Port))
		vMap["peerId"] = PeerID(addr.PeerID).String()
		arr[i] = vMap
	}
	return arr
}

func (x PeerRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}
