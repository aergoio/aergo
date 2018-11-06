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
