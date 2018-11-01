/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestBlockRequestHandler_handle(t *testing.T) {
	bigHash := make([]byte,6*1024*1024)
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
	}{
		{"TSingle", 1, 1, 1},
		{"TNotFounds", 10, 0, 1},
		{"TFound10", 100, 10, 10},
		{"TFoundAll", 20, 100, 20},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockMF := &v030MOFactory{}
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			//mockMF.On("newMsgResponseOrder", mock.Anything, GetBlocksResponse, mock.AnythingOfType("*types.GetBlockResponse")).Return(dummyMO)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("ID").Return(dummyPeerID)
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
			dummyMsg := &V030Message{id:NewMsgID()}
			msgBody := &types.GetBlockRequest{Hashes:make([][]byte, test.hashCnt)}
			h.handle(dummyMsg, msgBody)

			mockPeer.AssertNumberOfCalls(t, "sendAndWaitMessage", test.expectedSendCount)
		})
	}
}

func TestGetMissingHandler_sendMissingResp(t *testing.T) {
	bigHash := make([]byte, 2*1024*1024)
	logger := log.NewLogger("test")
	//validSmallBlockRsp := &message.GetBlockRsp{Block:&types.Block{Hash:make([]byte,40)},Err:nil}
	validBigBlockRsp := message.GetBlockRsp{Block:&types.Block{Hash:bigHash},Err:nil}
	notExistBlockRsp := message.GetBlockRsp{Block:nil,Err:nil}
	//dummyMO := new(MockMsgOrder)
	tests := []struct {
		name string
		hashCnt int
		validCallCount int
		expectedSendCount int
	}{
		{"TEmpty", 0, 0, 0},
		{"TSingle", 1, 1, 1},
		{"TActorError", 10, 0, 0},
		// max message size is clipped
		{"TFound10", 10, 10, 4},
		{"TFoundAll", 20, 20, 7},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPM := new(MockPeerManager)
			mockMF := &v030MOFactory{}
			mockPeer := new(MockRemotePeer)
			mockActor := new(MockActorService)
			//mockMF.On("newMsgResponseOrder", mock.Anything, GetBlocksResponse, mock.AnythingOfType("*types.GetBlockResponse")).Return(dummyMO)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("sendAndWaitMessage", mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)
			callReqCount :=0
			mockActor.On("CallRequestDefaultTimeout",message.ChainSvc, mock.MatchedBy(func(arg *message.GetBlockByNo) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return true
				}
				return false
			})).Return(validBigBlockRsp, nil)
			mockActor.On("CallRequestDefaultTimeout",message.ChainSvc, mock.MatchedBy(func(arg *message.GetBlockByNo) bool{
				callReqCount++
				if callReqCount <= test.validCallCount {
					return false
				}
				return true
			})).Return(notExistBlockRsp, nil)

			h:=newGetMissingReqHandler(mockPM, mockPeer, logger, mockActor)
			input := &message.GetMissingRsp{dummyTxHash, uint64(10), uint64(10+test.hashCnt)}
			h.sendMissingResp(mockPeer, NewMsgID(), input)
			//mockActor.AssertNumberOfCalls(t, "CallRequest", test.hashCnt)
			mockPeer.AssertNumberOfCalls(t, "sendAndWaitMessage", test.expectedSendCount)
		})
	}
}
