/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestBlockHashByNoReceiver_StartGet(t *testing.T) {
	inputNo := types.BlockNo(2222)
	tests := []struct {
		name string
		input types.BlockNo
		ttl time.Duration
	}{
		{"TSimple", inputNo, time.Millisecond*10},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.P2PSvc, mock.AnythingOfType("*types.GetHashByNo"))
			mockActor.On("TellRequest", message.SyncerSvc, mock.AnythingOfType("*message.GetHashByNoRsp"))
			mockMF := new(MockMoFactory)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)
			mockMF.On("newMsgBlockRequestOrder",mock.Anything, mock.Anything, mock.Anything).Return(dummyMo)

			expire := time.Now().Add(test.ttl)
			br := NewBlockHashByNoReceiver(mockActor, mockPeer, test.input, test.ttl)

			br.StartGet()

			assert.Equal(t, test.input, br.blockNo )
			assert.False(t, expire.After(br.timeout))

			// getBlock must be sent
			mockPeer.AssertCalled(t, "sendMessage", dummyMo)
			mockPeer.AssertNumberOfCalls(t, "sendMessage", 1)
		})
	}
}

func TestBlockHashByNoReceiver_ReceiveResp(t *testing.T) {
	blkNo := types.BlockNo(2222)
	blkHash := dummyBlockHash
	tests := []struct {
		name        string
		input       types.BlockNo
		ttl         time.Duration
		blkInterval time.Duration
		blkRsp      []byte
		rspStatus   types.ResultStatus

		// to verify
		consumed int
		sentResp int
		respError bool
	}{
		{"TSingleResp", blkNo, time.Minute, 0,  blkHash, types.ResultStatus_OK, 1,1, false},
		// Fail1 remote err
		{"TRemoteFail", blkNo, time.Minute,0,nil, types.ResultStatus_INTERNAL, 1,1, true},
		// Fail2 can't find block
		{"TMissingBlk", blkNo, time.Minute,0,nil,types.ResultStatus_NOT_FOUND,1,1, true},
		// Fail4 response sent after timeout
		{"TTimeout", blkNo, time.Millisecond*10,time.Millisecond*20,blkHash, types.ResultStatus_OK,1,0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.P2PSvc, mock.AnythingOfType("*types.GetHashByNo"))
			mockActor.On("TellRequest", message.SyncerSvc, mock.AnythingOfType("*message.GetHashByNoRsp"))
			//mockContext.On("Respond",mock.AnythingOfType("*message.GetBlockChunksRsp"))
			mockMF := new(MockMoFactory)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)
			mockPeer.On("consumeRequest", mock.AnythingOfType("p2pcommon.MsgID"))
			mockMF.On("newMsgBlockRequestOrder",mock.Anything, mock.Anything, mock.Anything).Return(dummyMo)

			//expire := time.Now().Add(test.ttl)
			br := NewBlockHashByNoReceiver(mockActor, mockPeer, test.input, test.ttl)
			br.StartGet()

			msg := &V030Message{subProtocol:GetHashByNoResponse, id: sampleMsgID}
			body := &types.GetHashByNoResponse{BlockHash:test.blkRsp, Status: test.rspStatus}
			if test.blkInterval > 0 {
				time.Sleep(test.blkInterval)
			}
			br.ReceiveResp(msg, body)

			mockPeer.AssertNumberOfCalls(t,"consumeRequest", test.consumed)
			mockActor.AssertNumberOfCalls(t, "TellRequest", test.sentResp)
			if test.sentResp > 0 {
				mockActor.AssertCalled(t, "TellRequest", message.SyncerSvc, mock.MatchedBy(func(arg *message.GetHashByNoRsp) bool {
					return (arg.Err != nil) == test.respError
				}))
			}
		})
	}
}
