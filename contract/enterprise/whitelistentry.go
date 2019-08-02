/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package enterprise

import (
	"errors"
	"github.com/aergoio/aergo/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/types"
	"net"
)

var notSpecifiedIP, notSpecifiedCIDR, _ = net.ParseCIDR("0.0.0.0/32")

const NotSpecifiedID = types.PeerID("")

var (
	InvalidEntryErr     = errors.New("invalid entry format")
	InvalidPeerIDErr    = errors.New("invalid peerID format")
	InvalidAddrRangeErr = errors.New("invalid address format")
	InvalidStateErr     = errors.New("either one of address or cidr is allowed")
)

type WhiteListEntry struct {
	literal string
	IpNet   *net.IPNet
	PeerID  types.PeerID
}

type rawEntry struct {
	Address string `json:"address"`
	Cidr    string `json:"cidr"`
	PeerId  string `json:"peerid"`
}

func NewWhiteListEntry(str string) (WhiteListEntry, error) {
	entry := WhiteListEntry{str, notSpecifiedCIDR, NotSpecifiedID}
	raw := rawEntry{}
	err := json.Unmarshal([]byte(str), &raw)
	if err != nil {
		return entry, InvalidEntryErr
	}
	if len(raw.Address) == 0 && len(raw.Cidr) == 0 && len(raw.PeerId) == 0 {
		return entry, InvalidEntryErr
	}
	if len(raw.Address) > 0 && len(raw.Cidr) > 0 {
		return entry, InvalidStateErr
	}
	if len(raw.PeerId) > 0 {
		pid, err := types.IDB58Decode(raw.PeerId)

		if err != nil || pid.Validate() != nil {
			return entry, InvalidPeerIDErr
		}
		entry.PeerID = pid
	}
	cidrStr := ""
	if len(raw.Address) > 0 {
		ip := net.ParseIP(raw.Address)
		if ip == nil {
			return entry, InvalidAddrRangeErr
		}
		if ip.To4() != nil {
			cidrStr = raw.Address + "/32"
		} else {
			cidrStr = raw.Address + "/128"
		}
	} else if len(raw.Cidr) > 0 {
		cidrStr = raw.Cidr
	}
	if len(cidrStr) > 0 {
		_, network, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return entry, InvalidAddrRangeErr
		}
		entry.IpNet = network
	}

	return entry, nil
}

func (e WhiteListEntry) Contains(addr net.IP, pid types.PeerID) bool {
	return e.checkAddr(addr) && e.checkPeerID(pid)
}

func (e WhiteListEntry) checkAddr(addr net.IP) bool {
	if e.IpNet != notSpecifiedCIDR {
		return e.IpNet.Contains(addr)
	} else {
		return true
	}
}
func (e WhiteListEntry) checkPeerID(pid types.PeerID) bool {
	if e.PeerID == NotSpecifiedID {
		return true
	} else {
		return pid == e.PeerID
	}
}

func (e WhiteListEntry) String() string {
	return e.literal
}
