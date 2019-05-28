/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/aergoio/aergo/p2p/p2pcommon"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
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
	LogBlkHash    = "blk_hash"
	LogBlkCount   = "blk_cnt"
	LogTxHash     = "tx_hash"
	LogTxCount    = "tx_cnt"
	LogRespStatus = "resp_st"
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
	var rsponse *message.MemPoolGetRsp
	switch v := rawResponse.(type) {
	case *message.MemPoolGetRsp:
		rsponse = v
	case message.MemPoolGetRsp:
		rsponse = &v
	default:
		panic("unexpected data type " + reflect.TypeOf(rawResponse).Name() + "is passed. check if there is a bug. ")
	}
	return extractTXs(rsponse)
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
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("no external ip address found")
}

func DebugLogReceiveMsg(logger *log.Logger, protocol p2pcommon.SubProtocol, msgID string, peer p2pcommon.RemotePeer, additional interface{}) {
	if additional != nil {
		logger.Debug().Str(LogProtoID, protocol.String()).Str(LogMsgID, msgID).Str("from_peer", peer.Name()).Str("other", fmt.Sprint(additional)).
			Msg("Received a message")
	} else {
		logger.Debug().Str(LogProtoID, protocol.String()).Str(LogMsgID, msgID).Str("from_peer", peer.Name()).
			Msg("Received a message")
	}
}

func DebugLogReceiveResponseMsg(logger *log.Logger, protocol p2pcommon.SubProtocol, msgID string, reqID string, peer p2pcommon.RemotePeer, additional interface{}) {
	if additional != nil {
		logger.Debug().Str(LogProtoID, protocol.String()).Str(LogMsgID, msgID).Str(LogOrgReqID, reqID).Str("from_peer", peer.Name()).Str("other", fmt.Sprint(additional)).
			Msg("Received a response message")
	} else {
		logger.Debug().Str(LogProtoID, protocol.String()).Str(LogMsgID, msgID).Str(LogOrgReqID, reqID).Str("from_peer", peer.Name()).
			Msg("Received a response message")
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

// bytesArrToString converts array of byte array to json array of b58 encoded string.
func BytesArrToString(bbarray [][]byte) string {
	return bytesArrToStringWithLimit(bbarray, 10)
}

func bytesArrToStringWithLimit(bbarray [][]byte, limit int) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	var arrSize = len(bbarray)
	if limit > arrSize {
		limit = arrSize
	}
	for i := 0; i < limit; i++ {
		hash := bbarray[i]
		buf.WriteByte('"')
		buf.WriteString(enc.ToString(hash))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	if arrSize > limit {
		buf.WriteString(fmt.Sprintf(" (and %d more), ", arrSize-limit))
	}
	buf.WriteByte(']')
	return buf.String()
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
