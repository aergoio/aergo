/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/satori/go.uuid"
)


type v030MOFactory struct {
}

func (mf *v030MOFactory) newMsgRequestOrder(expectResponse bool, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgBlockRequestOrder(respReceiver ResponseReceiver, protocolID SubProtocol, message pbMessage) msgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		rmo.respReceiver = respReceiver
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newMsgResponseOrder(reqID MsgID, protocolID SubProtocol, message pbMessage) msgOrder {
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

func (mf *v030MOFactory) newHandshakeMessage(protocolID SubProtocol, message pbMessage) Message {
	// TODO define handshake specific datatype
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo.message
	}
	return nil
}

// newPbMsgOrder is base form of making sendrequest struct
// TODO: It seems to have redundant parameter. reqID, expecteResponse and gossip param seems to be compacted to one or two parameters.
func newV030MsgOrder(mo *pbMessageOrder, msgID, orgID uuid.UUID, protocolID SubProtocol, message pbMessage) bool {
	bytes, err := marshalMessage(message)
	if err != nil {
		return false
	}

	var id, originalid MsgID
	copy(id[:],msgID[:])
	copy(originalid[:],orgID[:])

	msg := &V030Message{id: id, originalID:originalid,timestamp:time.Now().Unix(), subProtocol:protocolID,payload:bytes,length:uint32(len(bytes))}
	mo.protocolID = protocolID
	mo.needSign = true
	mo.message = msg

	return true
}
