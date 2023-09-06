/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/types"
	"github.com/gofrs/uuid"
)

// frequently used constants for indicating p2p log category
const (
	LogPeerID     = "peer_id"
	LogFullID     = "full_id" // LogFullID is Full qualified peer id
	LogPeerName   = "peer_nm"
	LogProtoID    = "protocol_id"
	LogMsgID      = "msg_id"
	LogOrgReqID   = "req_id" // LogOrgReqID is msgid of request from remote peer
	LogBlkHash    = types.LogBlkHash
	LogBlkNo      = types.LogBlkNo
	LogBlkCount   = "blk_cnt"
	LogTxHash     = "tx_hash"
	LogTxCount    = "tx_cnt"
	LogRespStatus = types.LogRespStatus
	LogRaftMsg    = "raftMsg"
)

func ExtractBlockFromRequest(rawResponse interface{}, err error) (*types.Block, error) {
	if err != nil {
		return nil, err
	}
	var blockRsp *message.GetBlockRsp
	switch v := rawResponse.(type) {
	case message.GetBlockRsp:
		blockRsp = &v
	case message.GetBestBlockRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	case message.GetBlockByNoRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	default:
		panic("unexpected data type " + reflect.TypeOf(rawResponse).Name() + "is passed. check if there is a bug. ")
	}
	return extractBlock(blockRsp)
}

func extractBlock(from *message.GetBlockRsp) (*types.Block, error) {
	if nil != from.Err {
		return nil, from.Err
	}
	return from.Block, nil

}

func extractTXsFromRequest(rawResponse interface{}, err error) ([]*types.Tx, error) {
	if err != nil {
		return nil, err
	}
	var response *message.MemPoolGetRsp
	switch v := rawResponse.(type) {
	case *message.MemPoolGetRsp:
		response = v
	case message.MemPoolGetRsp:
		response = &v
	default:
		panic("unexpected data type " + reflect.TypeOf(rawResponse).Name() + "is passed. check if there is a bug. ")
	}
	return extractTXs(response)
}

func extractTXs(from *message.MemPoolGetRsp) ([]*types.Tx, error) {
	if from.Err != nil {
		return nil, from.Err
	}
	txs := make([]*types.Tx, 0)
	for _, x := range from.Txs {
		txs = append(txs, x.GetTx())
	}
	return txs, nil
}

func setIP(a *types.PeerAddress, ipAddress net.IP) {
	a.Address = ipAddress.String()
}

// RandomUUID generate random UUID and return in form of string
func RandomUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func ExternalIP() (net.IP, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifs {
		if (i.Flags & net.FlagUp) == 0 {
			continue // downed
		}
		if (i.Flags & net.FlagLoopback) != 0 {
			continue // loopback
		}
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		ip := getValidIP(addrs)
		if ip != nil {
			return ip, err
		}
	}
	return nil, errors.New("no external ip address found")
}

func getValidIP(addrs []net.Addr) net.IP {
	var validIP6 net.IP = nil
	for _, addr := range addrs {
		ip := extractIP(addr)
		if ip == nil || ip.IsLoopback() {
			continue
		}
		//Drop link local address
		if ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
			continue
		}
		if ip.To4() != nil {
			return ip
		} else {
			validIP6 = ip
		}
	}
	if validIP6 != nil {
		return validIP6
	}
	return nil
}

func extractIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP
	case *net.IPNet:
		return v.IP
	default:
		return nil
	}
}

// ComparePeerID do byte-wise compare of two peerIDs,
func ComparePeerID(pid1, pid2 types.PeerID) int {
	p1 := []byte(string(pid1))
	p2 := []byte(string(pid2))
	l1 := len(p1)
	l2 := len(p2)
	compLen := l1
	if l1 > l2 {
		compLen = l2
	}

	// check bytes
	for i := 0; i < compLen; i++ {
		if comp := p1[i] - p2[i]; comp != 0 {
			if (comp & 0x80) == 0 {
				return int(comp)
			}
			return -1
		}
	}
	// check which is longer
	return l1 - l2
}

func PrintHashList(blocks []*types.Block) string {
	l := len(blocks)
	switch l {
	case 0:
		return "blk_cnt=0"
	case 1:
		return fmt.Sprintf("blk_cnt=1,hash=%s(num %d)", enc.ToString(blocks[0].Hash), blocks[0].Header.BlockNo)
	default:
		return fmt.Sprintf("blk_cnt=%d,firstHash=%s(num %d),lastHash=%s(num %d)", l, enc.ToString(blocks[0].Hash), blocks[0].Header.BlockNo, enc.ToString(blocks[l-1].Hash), blocks[l-1].Header.BlockNo)
	}

}

func PrintChainID(id *types.ChainID) string {
	var b strings.Builder
	if id.PublicNet {
		b.WriteString("publ")
	} else {
		b.WriteString("priv")
	}
	b.WriteByte(':')
	if id.MainNet {
		b.WriteString("main")
	} else {
		b.WriteString("test")
	}
	b.WriteByte(':')
	b.WriteString(id.Magic)
	b.WriteByte(':')
	b.WriteString(id.Consensus)
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(int(id.Version)))
	return b.String()
}
