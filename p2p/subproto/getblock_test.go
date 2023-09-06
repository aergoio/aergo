/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlockRequestHandler_handle(t *testing.T) {
	logger := log.NewLogger("test.subproto")

	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")

	bigHash := make([]byte, 2*1024*1024)
	//validSmallBlockRsp := &message.GetBlockRsp{Block:&types.Block{Hash:make([]byte,40)},Err:nil}
	validBlock := &types.Block{Hash: bigHash}
	//validBigBlockRsp := message.GetBlockRsp{Block:validBlock,Err:nil}
	//notExistBlockRsp := message.GetBlockRsp{Block:nil,Err:nil}
	//dummyMO := p2pmock.NewMockMsgOrder(ctrl)
	tests := []struct {
		name              string
		hashCnt           int
		validCallCount    int
		expectedSendCount int
		succResult        bool
	}{
		{"TSingle", 1, 1, 1, true},
		// not found return err result (ResultStatus_NOT_FOUND)
		{"TNotFounds", 10, 0, 1, false},
		// 4 blocks can be send in single message
		{"TFound10", 100, 10, 3, false},
		{"TFoundAll", 21, 100, 6, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			// mockery Mock can't handle big byte slice sell. it takes to much time to do. so use dummy stub instead and give up verify code.
			mockMF := &testDoubleMOFactory{}
			mockPeer.EXPECT().MF().Return(mockMF).AnyTimes()
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().Name().Return("16..aadecf@1").AnyTimes()
			mockPeer.EXPECT().SendAndWaitMessage(gomock.Any(), gomock.AssignableToTypeOf(time.Duration(0))).Return(nil).Times(test.expectedSendCount)

			callReqCount := 0
			mockCA := p2pmock.NewMockChainAccessor(ctrl)
			mockActor.EXPECT().GetChainAccessor().Return(mockCA).AnyTimes()
			mockCA.EXPECT().GetBlock(gomock.Any()).DoAndReturn(func(blockHash []byte) (*types.Block, error) {
				callReqCount++
				if callReqCount <= test.validCallCount {
					return validBlock, nil
				}
				return nil, nil
			}).MinTimes(1)

			h := NewBlockReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg := &testMessage{subProtocol: p2pcommon.GetBlocksRequest, id: p2pcommon.NewMsgID()}
			msgBody := &types.GetBlockRequest{Hashes: make([][]byte, test.hashCnt)}
			//h.Handle(dummyMsg, msgBody)
			h.handleBlkReq(dummyMsg, msgBody)

			// wait to work finished
			<-h.w

			mockMF.mutex.Lock()
			lastStatus := mockMF.lastStatus
			mockMF.mutex.Unlock()

			assert.Equal(t, test.succResult, lastStatus == types.ResultStatus_OK)
		})
	}
}

func TestBlockResponseHandler_handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := log.NewLogger("test.subproto")
	var dummyPeerID, _ = types.IDB58Decode("16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD")
	dummyBlockHash, _ := enc.ToBytes("v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6")
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

	blkNo := uint64(100)
	prevHash := dummyBlockHash
	inputHashes := make([]message.BlockHash, len(sampleBlks))
	inputBlocks := make([]*types.Block, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = hash
		inputBlocks[i] = &types.Block{Hash: hash, Header: &types.BlockHeader{PrevBlockHash: prevHash, BlockNo: blkNo}}
		blkNo++
		prevHash = hash
	}

	tests := []struct {
		name string

		receiver    p2pcommon.ResponseReceiver
		wantConsume int
		wantCallSM  int
	}{
		// 1. not exist receiver and consumed message
		//{"Tnothing",nil, true},
		// 2. exist receiver and consume successfully
		{"TexistAndConsume", func(msg p2pcommon.Message, body p2pcommon.MessageBody) bool {
			return true
		}, 0, 0},
		// 2. exist receiver but not consumed
		{"TExistWrong", func(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) bool {
			return false
		}, 1, 1},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().ConsumeRequest(gomock.AssignableToTypeOf(p2pcommon.MsgID{})).Times(test.wantConsume)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockSM := p2pmock.NewMockSyncManager(ctrl)
			mockSM.EXPECT().HandleGetBlockResponse(mockPeer, gomock.Any(), gomock.AssignableToTypeOf(&types.GetBlockResponse{})).Times(test.wantCallSM)

			mockPeer.EXPECT().GetReceiver(gomock.AssignableToTypeOf(p2pcommon.MsgID{})).Return(test.receiver)
			msg := &testMessage{subProtocol: p2pcommon.GetBlocksResponse, id: p2pcommon.NewMsgID()}
			body := &types.GetBlockResponse{Blocks: make([]*types.Block, 2)}
			h := NewBlockRespHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			h.Handle(msg, body)
		})
	}
}
