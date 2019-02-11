/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
)


type v030MOFactory struct {
}

func (mf *v030MOFactory) newMsgRequestOrder(expectResponse bool, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		rmo.respReceiver = respReceiver
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	rmo := &pbResponseOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.FromBytesOrNil(reqID[:]), protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) msgOrder {
	rmo := &pbBlkNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, NewBlockNotice, noticeMsg) {
		rmo.blkHash = noticeMsg.BlockHash
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgTxBroadcastOrder(message *types.NewTransactionsNotice) msgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, reqID, uuid.Nil, NewTxNotice, message) {
		rmo.txHashes = message.TxHashes
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) msgOrder {
	rmo := &pbBpNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, BlockProducedNotice, noticeMsg) {
		rmo.block = noticeMsg.Block
		return rmo
	}
	return nil
}


func (mf *v030MOFactory) newHandshakeMessage(protocolID p2pcommon.SubProtocol, message pbMessage) p2pcommon.Message {
	// TODO define handshake specific datatype
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo.message
	}
	return nil
}

// newPbMsgOrder is base form of making sendrequest struct
func newV030MsgOrder(mo *pbMessageOrder, msgID, orgID uuid.UUID, protocolID p2pcommon.SubProtocol, message pbMessage) bool {
	bytes, err := MarshalMessage(message)
	if err != nil {
		return false
	}

	var id, originalid p2pcommon.MsgID
	copy(id[:],msgID[:])
	copy(originalid[:],orgID[:])

	msg := &V030Message{id: id, originalID:originalid,timestamp:time.Now().UnixNano(), subProtocol:protocolID,payload:bytes,length:uint32(len(bytes))}
	mo.protocolID = protocolID
	mo.needSign = true
	mo.message = msg

	return true
}
