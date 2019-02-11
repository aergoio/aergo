/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)


func TestBlockRequestHandler_handle(t *testing.T) {
	bigHash := make([]byte,2*1024*1024)
	logger := log.NewLogger("test")
	//validSmallBlockRsp := &message.GetBlockRsp{Block:&types.Block{Hash:make([]byte,40)},Err:nil}
	validBlock:= &types.Block{Hash:bigHash}
	//validBigBlockRsp := message.GetBlockRsp{Block:validBlock,Err:nil}
	//notExistBlockRsp := message.GetBlockRsp{Block:nil,Err:nil}
	//dummyMO := new(MockMsgOrder)
	tests := []struct {
		name string
		hashCnt int
		validCallCount int
		expectedSendCount int
		succResult bool
	}{
		{"TSingle", 1, 1, 1, true},
		// not found return err result (ResultStatus_NOT_FOUND)
		{"TNotFounds", 10, 0, 1, false},
		{"TFound10", 100, 10, 4, false},
		{"TFoundAll", 20, 100, 7, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			// mockery Mock can't handle big byte slice sell. it takes to much time to do. so use dummy stub instead and give up verify code.
			mockMF := &testDoubleFactory{}
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("Name").Return("16..aadecf@1")
			mockPeer.On("sendAndWaitMessage", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

			callReqCount :=0
			mockCA := new(MockChainAccessor)
			mockActor.On("GetChainAccessor").Return(mockCA)
			mockCA.On("GetBlock", mock.MatchedBy(func(arg []byte) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return true
				}
				return false
			})).Return(validBlock, nil)
			mockCA.On("GetBlock", mock.MatchedBy(func(arg []byte) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return false
				}
				return true
			})).Return(nil, nil)

			h:=newBlockReqHandler(mockPM, mockPeer, logger, mockActor)
			dummyMsg := &V030Message{id: p2pcommon.NewMsgID()}
			msgBody := &types.GetBlockRequest{Hashes:make([][]byte, test.hashCnt)}
			h.handle(dummyMsg, msgBody)

			mockPeer.AssertNumberOfCalls(t, "sendAndWaitMessage", test.expectedSendCount)
			assert.Equal(t, test.succResult, mockMF.lastStatus == types.ResultStatus_OK )
		})
	}
}

type testDoubleFactory struct {
	v030MOFactory
	lastStatus types.ResultStatus
}
func (f *testDoubleFactory) newMsgResponseOrder(reqID p2pcommon.MsgID, protocolID p2pcommon.SubProtocol, message pbMessage) msgOrder {
	f.lastStatus = message.(*types.GetBlockResponse).Status
	return f.v030MOFactory.newMsgResponseOrder(reqID, protocolID, message)
}

func TestBlockResponseHandler_handle(t *testing.T) {
	blkNo  := uint64(100)
	prevHash := dummyBlockHash
	inputHashes := make([]message.BlockHash,len(sampleBlks))
	inputBlocks := make([]*types.Block,len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = hash
		inputBlocks[i] = &types.Block{Hash:hash, Header:&types.BlockHeader{PrevBlockHash:prevHash, BlockNo:blkNo}}
		blkNo++
		prevHash = hash
	}

	tests := []struct {
		name string

		receiver ResponseReceiver
		consume bool
		callSM bool
	}{
		// 1. not exist receiver and consumed message
		//{"Tnothing",nil, true},
		// 2. exist receiver and consume successfully
		{"TexistAndConsume", func(msg p2pcommon.Message, body proto.Message) bool {
			return true
		}, true, false},
		// 2. exist receiver but not consumed
		{"TExistWrong", dummyResponseReceiver, false, true},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("consumeRequest", mock.AnythingOfType("p2pcommon.MsgID"))
			mockActor := new(MockActorService)
			mockSM := new(MockSyncManager)
			mockSM.On("HandleGetBlockResponse",mockPeer, mock.Anything, mock.AnythingOfType("*types.GetBlockResponse"))

			mockPeer.On("GetReceiver", mock.AnythingOfType("p2pcommon.MsgID")).Return(test.receiver)
			msg := &V030Message{subProtocol:GetBlocksResponse, id: sampleMsgID}
			body := &types.GetBlockResponse{Blocks:make([]*types.Block,2)}
			h := newBlockRespHandler(mockPM, mockPeer, logger, mockActor, mockSM)
			h.handle(msg, body)
			if  test.consume {
				mockSM.AssertNumberOfCalls(t, "HandleGetBlockResponse", 0)
			} else {
				mockSM.AssertNumberOfCalls(t, "HandleGetBlockResponse", 1)
			}
		})
	}
}
