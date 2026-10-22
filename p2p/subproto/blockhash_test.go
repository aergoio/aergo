/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"bytes"
	"crypto/sha256"
	"sync"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetHashRequestHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")

	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var sampleMsgID = p2pcommon.NewMsgID()

	var sampleBlksB58 = []string{
		"v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6",
		"2VEPg4MqJUoaS3EhZ6WWSAUuFSuD4oSJ645kSQsGV7H9",
		"AtzTZ2CZS45F1276RpTdLfYu2DLgRcd9HL3aLqDT1qte",
		"2n9QWNDoUvML756X7xdHWCFLZrM4CQEtnVH2RzG5FYAw",
		"6cy7U7XKYtDTMnF3jNkcJvJN5Rn85771NSKjc5Tfo2DM",
		"3bmB8D37XZr4DNPs64NiGRa2Vw3i8VEgEy6Xc2XBmRXC",
	}
	var sampleBlks [][]byte
	var sampleBlksHashes []types.BlockID

	sampleBlks = make([][]byte, len(sampleBlksB58))
	sampleBlksHashes = make([]types.BlockID, len(sampleBlksB58))
	for i, hashb58 := range sampleBlksB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleBlks[i] = hash
		copy(sampleBlksHashes[i][:], hash)
	}

	baseHeight := uint64(110000)
	sampleSize := 21
	mainChainHashes := make([][]byte, sampleSize)
	sideChainHashes := make([][]byte, sampleSize)
	digest := sha256.New()
	for i := 0; i < sampleSize; i++ {
		digest.Write(uuid.Must(uuid.NewV4()).Bytes())
		mainChainHashes[i] = digest.Sum(nil)
		digest.Write(uuid.Must(uuid.NewV4()).Bytes())
		sideChainHashes[i] = digest.Sum(nil)
	}
	tests := []struct {
		name   string
		inNum  uint64
		inHash []byte
		inSize uint64

		firstChain [][]byte
		reorgIdx   int
		lastChain  [][]byte

		expectedStatus  types.ResultStatus
		expectedHashCnt int
	}{
		// 1. success (exact prev and enough chaining)
		{"Tsucc", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_OK, 20},
		// 2. exact prev but smaller chaining
		{"TShorter", baseHeight, mainChainHashes[0], 20,
			mainChainHashes[:16], 99999, mainChainHashes[:16], types.ResultStatus_OK, 15},
		// 3. wrong prev
		{"TWrongPrev", baseHeight, sampleBlks[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_INVALID_ARGUMENT, 0},
		// 4. missing prev (smaller best block than prev)
		{"TMissingPrev", baseHeight + 30, mainChainHashes[0], 20,
			mainChainHashes, 99999, mainChainHashes, types.ResultStatus_INVALID_ARGUMENT, 0},
		// 5. exact prev , but reorg from middle before first fetch
		{"TReorgBefore", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 0, append(append(make([][]byte, 0, sampleSize), mainChainHashes[:5]...), sideChainHashes[5:]...), types.ResultStatus_OK, 20},
		// 6. exact prev , but changed (by such as reorg) during fetch
		{"TReorgMid", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 10, append(append(make([][]byte, 0, sampleSize), mainChainHashes[:5]...), sideChainHashes[5:]...), types.ResultStatus_INTERNAL, 0},
		// 7. exact prev at first, but changed prev (and decent blocks also) before first fetch
		{"TReorgWhole", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 0, sideChainHashes, types.ResultStatus_INTERNAL, 0},
		// 7. exact prev at first, but changed prev (and decent blocks also) during fetch
		{"TReorgWhole2", baseHeight, mainChainHashes[0], 20,
			mainChainHashes, 10, sideChainHashes, types.ResultStatus_INTERNAL, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			dummyMF := &testDoubleHashesRespFactory{}
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().MF().Return(dummyMF).MinTimes(1)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

			mockAcc := &testDoubleChainAccessor{firstChain: test.firstChain, lastChain: test.lastChain, baseHeight: baseHeight, reorgTiming: test.reorgIdx}
			mockActor.EXPECT().GetChainAccessor().Return(mockAcc).MinTimes(1)

			msg := p2pmock.NewMockMessage(ctrl)
			msg.EXPECT().ID().Return(sampleMsgID).AnyTimes()
			msg.EXPECT().Subprotocol().Return(p2pcommon.GetHashesRequest).AnyTimes()
			body := &types.GetHashesRequest{PrevNumber: test.inNum, PrevHash: test.inHash, Size: test.inSize}

			h := NewGetHashesReqHandler(mockPM, mockPeer, logger, mockActor)
			h.Handle(msg, body)

			// verify whether handler send response with expected result
			assert.Equal(t, test.expectedStatus.String(), dummyMF.lastStatus.String())
			if test.expectedStatus == types.ResultStatus_OK {
				assert.Equal(t, test.expectedHashCnt, len(dummyMF.lastResp.Hashes))
			}
		})
	}
}

func TestGetHashByNoRequestHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")

	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	var sampleMsgID = p2pcommon.NewMsgID()

	var sampleBlksB58 = []string{
		"v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6",
		"2VEPg4MqJUoaS3EhZ6WWSAUuFSuD4oSJ645kSQsGV7H9",
		"AtzTZ2CZS45F1276RpTdLfYu2DLgRcd9HL3aLqDT1qte",
		"2n9QWNDoUvML756X7xdHWCFLZrM4CQEtnVH2RzG5FYAw",
		"6cy7U7XKYtDTMnF3jNkcJvJN5Rn85771NSKjc5Tfo2DM",
		"3bmB8D37XZr4DNPs64NiGRa2Vw3i8VEgEy6Xc2XBmRXC",
	}
	var sampleBlks [][]byte
	var sampleBlksHashes []types.BlockID

	sampleBlks = make([][]byte, len(sampleBlksB58))
	sampleBlksHashes = make([]types.BlockID, len(sampleBlksB58))
	for i, hashb58 := range sampleBlksB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleBlks[i] = hash
		copy(sampleBlksHashes[i][:], hash)
	}

	baseHeight := uint64(110000)
	wrongHeight := uint64(21531535)
	tests := []struct {
		name      string
		inNum     uint64
		accRet    []byte
		accRetErr error

		expectedStatus types.ResultStatus
	}{
		// 1. success (exact prev and enough chaining)
		{"Tsucc", baseHeight, sampleBlks[0], nil, types.ResultStatus_OK},
		// 2. exact prev but smaller chaining
		{"TMissing", wrongHeight, nil, &chain.ErrNoBlock{}, types.ResultStatus_NOT_FOUND},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			dummyMF := &testDoubleMOFactory{}
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().MF().Return(dummyMF).MinTimes(1)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

			mockAcc := p2pmock.NewMockChainAccessor(ctrl)
			mockAcc.EXPECT().GetHashByNo(test.inNum).Return(test.accRet, test.accRetErr).Times(1)
			mockActor.EXPECT().GetChainAccessor().Return(mockAcc).AnyTimes()

			msg := p2pmock.NewMockMessage(ctrl)
			msg.EXPECT().ID().Return(sampleMsgID).AnyTimes()
			msg.EXPECT().Subprotocol().Return(p2pcommon.GetHashByNoRequest).AnyTimes()
			body := &types.GetHashByNo{BlockNo: test.inNum}

			h := NewGetHashByNoReqHandler(mockPM, mockPeer, logger, mockActor)
			h.Handle(msg, body)

			// verify
			assert.Equal(t, test.expectedStatus.String(), dummyMF.lastStatus.String())
		})
	}
}

func (a *testDoubleChainAccessor) GetBestBlock() (*types.Block, error) {
	mychain := a.getChain()
	idx := len(mychain) - 1
	return &types.Block{Hash: mychain[idx], Header: &types.BlockHeader{BlockNo: a.baseHeight + types.BlockNo(idx)}}, nil
}

func (a *testDoubleChainAccessor) GetConsensusInfo() string {
	return ""
}

// GetBlock return block of blockHash. It return nil and error if not found block of that hash or there is a problem in db store
func (a *testDoubleChainAccessor) GetBlock(blockHash []byte) (*types.Block, error) {
	mychain := a.getChain()
	for i, hash := range mychain {
		if bytes.Equal(hash, blockHash) {
			prevHash := []byte(nil)
			if i > 0 {
				prevHash = mychain[i-1]
			}
			number := a.baseHeight + types.BlockNo(i)
			return &types.Block{Hash: hash, Header: &types.BlockHeader{BlockNo: number, PrevBlockHash: prevHash}}, nil
		}
	}
	return nil, chain.ErrNoBlock{}
}

// GetHashByNo returns hash of block. It return nil and error if not found block of that number or there is a problem in db store
func (a *testDoubleChainAccessor) GetHashByNo(blockNo types.BlockNo) ([]byte, error) {
	mychain := a.getChain()
	a.callCount++
	idx := blockNo - a.baseHeight
	if idx < 0 || idx >= types.BlockNo(len(mychain)) {
		return nil, chain.ErrNoBlock{}
	}
	return mychain[int(idx)], nil
}

// emulate db
type testDoubleChainAccessor struct {
	p2pmock.MockChainAccessor
	baseHeight types.BlockNo

	callCount   int
	reorgTiming int
	firstChain  [][]byte
	lastChain   [][]byte
}

var _ types.ChainAccessor = (*testDoubleChainAccessor)(nil)

func (a *testDoubleChainAccessor) getChain() [][]byte {
	if a.callCount <= a.reorgTiming {
		return a.firstChain
	} else {
		return a.lastChain
	}
}

type testDoubleHashesRespFactory struct {
	lastResp   *types.GetHashesResponse
	lastStatus types.ResultStatus
}

func (f *testDoubleHashesRespFactory) NewMsgRequestOrder(expecteResponse bool, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewMsgRequestOrderWithReceiver(respReceiver p2pcommon.ResponseReceiver, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message p2pcommon.MessageBody) p2pcommon.MsgOrder {
	f.lastResp = message.(*types.GetHashesResponse)
	f.lastStatus = f.lastResp.Status
	return &testMo{message: &testMessage{id: reqID, subProtocol: protocolID}}
}

func (f *testDoubleHashesRespFactory) NewMsgBlkBroadcastOrder(noticeMsg *types.NewBlockNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewMsgTxBroadcastOrder(noticeMsg *types.NewTransactionsNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewMsgBPBroadcastOrder(noticeMsg *types.BlockProducedNotice) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewRaftMsgOrder(msgType raftpb.MessageType, raftMsg *raftpb.Message) p2pcommon.MsgOrder {
	panic("implement me")
}

func (f *testDoubleHashesRespFactory) NewTossMsgOrder(orgMsg p2pcommon.Message) p2pcommon.MsgOrder {
	panic("implement me")
}

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
	panic("implement me")
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
