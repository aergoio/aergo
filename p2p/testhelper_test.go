package p2p

import (
	"sync"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

// testDoubleMOFactory keep last created message and last result status of response message
type testDoubleMOFactory struct {
	mutex      sync.Mutex
	lastResp   p2pcommon.MessageBody
	lastStatus types.ResultStatus
}

func (f *testDoubleMOFactory) NewTossMsgOrder(orgMsg p2pcommon.Message) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleMOFactory) NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleMOFactory) NewMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleMOFactory) NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleMOFactory) NewMsgRequestOrder(expecteResponse bool, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleMOFactory) NewMsgRequestOrderWithReceiver(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.lastResp = message
	return &testMo{message: &testMessage{subProtocol: protocolID}}
}

func (f *testDoubleMOFactory) NewMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.lastResp = message
	f.lastStatus = f.lastResp.(types.ResponseMessage).GetStatus()
	return &testMo{message: &testMessage{id: reqID, subProtocol: protocolID}}
}

func (f *testDoubleMOFactory) NewRaftMsgOrder(msgType raftpb.MessageType, raftMsg *raftpb.Message) p2pcommon.MsgOrder {
	panic("implement me")
}

type testMo struct {
	protocolID p2pcommon.SubProtocol // protocolName and msg struct type MUST be matched.
	message    p2pcommon.Message
}

func (mo *testMo) GetMsgID() p2pcommon.MsgID {
	return mo.message.ID()
}

func (mo *testMo) Timestamp() int64 {
	return mo.message.Timestamp()
}

func (*testMo) IsRequest() bool {
	return false
}

func (*testMo) IsNeedSign() bool {
	return true
}

func (mo *testMo) GetProtocolID() p2pcommon.SubProtocol {
	return mo.protocolID
}

func (mo *testMo) SendTo(p p2pcommon.RemotePeer) error {
	return nil
}

func (mo *testMo) CancelSend(p p2pcommon.RemotePeer) {

}

type testMessage struct {
	subProtocol p2pcommon.SubProtocol
	// Length is lenght of payload
	length uint32
	// timestamp is unix time (precision of second)
	timestamp int64
	// ID is 16 bytes unique identifier
	id p2pcommon.MsgID
	// OriginalID is message id of request which trigger this message. it will be all zero, if message is request or notice.
	originalID p2pcommon.MsgID

	// marshaled by google protocol buffer v3. object is determined by Subprotocol
	payload []byte
}

func (m *testMessage) Subprotocol() p2pcommon.SubProtocol {
	return m.subProtocol
}

func (m *testMessage) Length() uint32 {
	return m.length

}

func (m *testMessage) Timestamp() int64 {
	return m.timestamp
}

func (m *testMessage) ID() p2pcommon.MsgID {
	return m.id
}

func (m *testMessage) OriginalID() p2pcommon.MsgID {
	return m.originalID
}

func (m *testMessage) Payload() []byte {
	return m.payload
}
