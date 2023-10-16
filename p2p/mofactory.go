/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/gofrs/uuid"
)

type baseMOFactory struct {
	is p2pcommon.InternalService
}

func (mf *baseMOFactory) NewMsgRequestOrder(expectResponse bool, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewMsgRequestOrderWithReceiver(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		rmo.respReceiver = respReceiver
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	rmo := &pbResponseOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.FromBytesOrNil(reqID[:]), protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) p2pcommon.MsgOrder {
	rmo := &pbBlkNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, p2pcommon.NewBlockNotice, noticeMsg) {
		rmo.blkHash = noticeMsg.BlockHash
		rmo.blkNo = noticeMsg.BlockNo
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewMsgTxBroadcastOrder(message *types.NewTransactionsNotice) p2pcommon.MsgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, reqID, uuid.Nil, p2pcommon.NewTxNotice, message) {
		rmo.txHashes = make([]types.TxID, len(message.TxHashes))
		for i, h := range message.TxHashes {
			rmo.txHashes[i] = types.ToTxID(h)
		}
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) p2pcommon.MsgOrder {
	rmo := &pbBpNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, p2pcommon.BlockProducedNotice, noticeMsg) {
		rmo.block = noticeMsg.Block
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewRaftMsgOrder(msgType raftpb.MessageType, raftMsg *raftpb.Message) p2pcommon.MsgOrder {
	rmo := &pbRaftMsgOrder{msg: raftMsg, raftAcc: mf.is.ConsensusAccessor().RaftAccessor()}
	msgID := uuid.Must(uuid.NewV4())
	if mf.fillUpMsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, p2pcommon.RaftWrapperMessage, raftMsg) {
		switch msgType {
		case raftpb.MsgHeartbeat, raftpb.MsgHeartbeatResp:
			rmo.trace = false
		default:
			// follow default policy
		}
		return rmo
	}
	return nil
}

func (mf *baseMOFactory) NewTossMsgOrder(orgMsg p2pcommon.Message) p2pcommon.MsgOrder {
	rmo := &pbTossOrder{pbMessageOrder{message: orgMsg, protocolID: orgMsg.Subprotocol(), needSign: true}}
	return rmo
}

// newPbMsgOrder is base form of making sendRequest struct
func (mf *baseMOFactory) fillUpMsgOrder(mo *pbMessageOrder, msgID, orgID uuid.UUID, protocolID p2pcommon.SubProtocol, messageBody p2pcommon.MessageBody) bool {
	id := p2pcommon.MsgID(msgID)
	originalID := p2pcommon.MsgID(orgID)
	bytes, err := p2putil.MarshalMessageBody(messageBody)
	if err != nil {
		return false
	}
	msg := p2pcommon.NewMessageValue(protocolID, id, originalID, time.Now().UnixNano(), bytes)
	mo.protocolID = protocolID
	mo.needSign = true
	mo.message = msg
	mo.trace = true

	return true
}
