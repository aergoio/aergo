/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package enterprise

import (
	"errors"
	"github.com/aergoio/aergo/types"
	"net"
	"strings"
)

var _, notSpecifiedAddr, _ = net.ParseCIDR("0.0.0.0/32")

const NotSpecifiedID = types.PeerID("")

var (
	InvalidEntryErr     = errors.New("invalid entry format")
	InvalidPeerIDErr    = errors.New("invalid peerID format")
	InvalidAddrRangeErr = errors.New("invalid address format")
)

type WhiteListEntry struct {
	literal string
	IpNet   *net.IPNet
	PeerID  types.PeerID
}

func NewWhiteListEntry(str string) (WhiteListEntry, error) {
	entry := WhiteListEntry{str, notSpecifiedAddr, NotSpecifiedID}
	strs := strings.SplitN(str, ":", 2)
	if len(strs) != 2 {
		return entry, InvalidEntryErr
	}
	if len(strs[0]) > 0 {
		pid, err := types.IDB58Decode(strs[0])

		if err != nil || pid.Validate() != nil {
			return entry, InvalidPeerIDErr
		}
		entry.PeerID = pid
	}
	if len(strs[1]) > 0 {
		ip := net.ParseIP(strs[1])
		if ip != nil {
			if ip.To4() != nil {
				strs[1] = strs[1] + "/32"
			} else {
				strs[1] = strs[1] + "/128"
			}
		}
		_, network, err := net.ParseCIDR(strs[1])
		if err != nil {
			return entry, InvalidAddrRangeErr
		}
		entry.IpNet = network
	}

	return entry, nil
}

func (e WhiteListEntry) Contains(addr net.IP, pid types.PeerID) bool {
	if e.IpNet == notSpecifiedAddr {
		if e.PeerID == NotSpecifiedID {
			return true
		} else {
			return pid == e.PeerID
		}
	} else if e.PeerID == NotSpecifiedID {
		return e.IpNet.Contains(addr)
	} else {
		return e.IpNet.Contains(addr) && (pid == e.PeerID)
	}
}

func (e WhiteListEntry) String() string {
	return e.literal
}