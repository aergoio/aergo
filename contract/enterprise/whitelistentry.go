/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package enterprise

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/types"
	"io"
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

type RawEntry struct {
	Address string `json:"address"`
	Cidr    string `json:"cidr"`
	PeerId  string `json:"peerid"`
}
var dummyListEntry WhiteListEntry
func init() {
	dummyListEntry = WhiteListEntry{"", notSpecifiedCIDR, NotSpecifiedID}
}
func NewWhiteListEntry(str string) (WhiteListEntry, error) {
	raw := RawEntry{}
	err := json.Unmarshal([]byte(str), &raw)
	if err != nil {
		return dummyListEntry, InvalidEntryErr
	}
	return NewListEntry(raw)
}

func NewListEntry(raw RawEntry) (WhiteListEntry, error) {
	literal, _ := json.Marshal(raw)

	if len(raw.Address) == 0 && len(raw.Cidr) == 0 && len(raw.PeerId) == 0 {
		return dummyListEntry, InvalidEntryErr
	}
	if len(raw.Address) > 0 && len(raw.Cidr) > 0 {
		return dummyListEntry, InvalidStateErr
	}

	entry := WhiteListEntry{string(literal), notSpecifiedCIDR, NotSpecifiedID}
	if len(raw.PeerId) > 0 {
		pid, err := types.IDB58Decode(raw.PeerId)

		if err != nil || pid.Validate() != nil {
			return dummyListEntry, InvalidPeerIDErr
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

func ReadEntries(jsonBytes []byte) ([]WhiteListEntry, error) {
	var list []RawEntry

	err := json.Unmarshal(jsonBytes, &list)
	if err != nil {
		return nil, err
	}
	eList := make([]WhiteListEntry,len(list))
	for i, r := range list {
		eList[i],err = NewListEntry(r)
		if err != nil {
			return nil, fmt.Errorf("line %v. error %s",i, err.Error())
		}
	}
	return eList, nil
}

func WriteEntries(entries []WhiteListEntry, wr io.Writer) error {
	rList := make([]RawEntry,len(entries))
	for i, e := range entries {
		r := RawEntry{}
		if e.PeerID != NotSpecifiedID {
			r.PeerId = types.IDB58Encode(e.PeerID)
		}
		if e.IpNet != nil && e.IpNet != notSpecifiedCIDR {
			if m, b := e.IpNet.Mask.Size() ; m == b {
				// single ip
				r.Address = e.IpNet.IP.String()
			} else {
				r.Cidr = e.IpNet.String()
			}
		}

		rList[i] = r
	}
	en := json.NewEncoder(wr)
	return en.Encode(rList)
}