/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

//go:generate mockgen -source=internalmsg.go -package=p2pmock -destination=../p2pmock/mock_msgorder.go

// MsgOrder is abstraction of information about the message that will be sent to peer.
// Some type of msgOrder, such as notice mo, should thread-safe and re-entrant
type MsgOrder interface {
	GetMsgID() MsgID
	// Timestamp is unit time value
	Timestamp() int64
	IsRequest() bool
	IsNeedSign() bool
	GetProtocolID() SubProtocol

	// SendTo send message to remote peer. it return err if write fails, or nil if write is successful or ignored.
	SendTo(p RemotePeer) error
	CancelSend(pi RemotePeer)
}

type MoFactory interface {
	NewMsgRequestOrder(expectResponse bool, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgRequestOrderWithReceiver(respReceiver ResponseReceiver, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgResponseOrder(reqID MsgID, protocolID SubProtocol, message MessageBody) MsgOrder
	NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) MsgOrder
	NewMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) MsgOrder
	NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) MsgOrder
	NewRaftMsgOrder(msgType raftpb.MessageType, raftMsg *raftpb.Message) MsgOrder
	NewTossMsgOrder(orgMsg Message) MsgOrder
}
