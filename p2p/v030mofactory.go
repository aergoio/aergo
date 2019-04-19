/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/subproto"

	"github.com/aergoio/aergo/types"
	"github.com/gofrs/uuid"
)

type v030MOFactory struct {
}

func (mf *v030MOFactory) NewMsgRequestOrder(expectResponse bool, protocolID p2pcommon.SubProtocol, message p2pcommon.PbMessage) p2pcommon.MsgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) NewMsgBlockRequestOrder(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.PbMessage) p2pcommon.MsgOrder {
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		rmo.respReceiver = respReceiver
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) NewMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.PbMessage) p2pcommon.MsgOrder {
	rmo := &pbResponseOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.FromBytesOrNil(reqID[:]), protocolID, message) {
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) p2pcommon.MsgOrder {
	rmo := &pbBlkNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, subproto.NewBlockNotice, noticeMsg) {
		rmo.blkHash = noticeMsg.BlockHash
		rmo.blkNo = noticeMsg.BlockNo
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) NewMsgTxBroadcastOrder(message *types.NewTransactionsNotice) p2pcommon.MsgOrder {
	rmo := &pbTxNoticeOrder{}
	reqID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, reqID, uuid.Nil, subproto.NewTxNotice, message) {
		rmo.txHashes = message.TxHashes
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) p2pcommon.MsgOrder {
	rmo := &pbBpNoticeOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, subproto.BlockProducedNotice, noticeMsg) {
		rmo.block = noticeMsg.Block
		return rmo
	}
	return nil
}

func (mf *v030MOFactory) newHandshakeMessage(protocolID p2pcommon.SubProtocol, message p2pcommon.PbMessage) p2pcommon.Message {
	// TODO define handshake specific datatype
	rmo := &pbRequestOrder{}
	msgID := uuid.Must(uuid.NewV4())
	if newV030MsgOrder(&rmo.pbMessageOrder, msgID, uuid.Nil, protocolID, message) {
		return rmo.message
	}
	return nil
}

// newPbMsgOrder is base form of making sendrequest struct
func newV030MsgOrder(mo *pbMessageOrder, msgID, orgID uuid.UUID, protocolID p2pcommon.SubProtocol, message p2pcommon.PbMessage) bool {
	bytes, err := p2putil.MarshalMessage(message)
	if err != nil {
		return false
	}

	var id, originalid p2pcommon.MsgID
	copy(id[:], msgID[:])
	copy(originalid[:], orgID[:])

	msg := &V030Message{id: id, originalID: originalid, timestamp: time.Now().UnixNano(), subProtocol: protocolID, payload: bytes, length: uint32(len(bytes))}
	mo.protocolID = protocolID
	mo.needSign = true
	mo.message = msg

	return true
}
