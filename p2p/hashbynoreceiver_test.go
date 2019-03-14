/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pmocks"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlockHashByNoReceiver_StartGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inputNo := types.BlockNo(2222)
	tests := []struct {
		name  string
		input types.BlockNo
		ttl   time.Duration
	}{
		{"TSimple", inputNo, time.Millisecond * 10},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmocks.NewMockActorService(ctrl)

			mockMo := createDummyMo(ctrl)
			mockMF := p2pmocks.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(mockMo).Times(1)

			expire := time.Now().Add(test.ttl)
			br := NewBlockHashByNoReceiver(mockActor, mockPeer, test.input, test.ttl)

			br.StartGet()

			assert.Equal(t, test.input, br.blockNo)
			assert.False(t, expire.After(br.timeout))

			// getBlock must be sent
		})
	}
}

func TestBlockHashByNoReceiver_ReceiveResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
		consumed  int
		sentResp  int
		respError bool
	}{
		{"TSingleResp", blkNo, time.Minute, 0, blkHash, types.ResultStatus_OK, 1, 1, false},
		// Fail1 remote err
		{"TRemoteFail", blkNo, time.Minute, 0, nil, types.ResultStatus_INTERNAL, 1, 1, true},
		// Fail2 can't find block
		{"TMissingBlk", blkNo, time.Minute, 0, nil, types.ResultStatus_NOT_FOUND, 1, 1, true},
		// Fail4 response sent after timeout
		{"TTimeout", blkNo, time.Millisecond * 10, time.Millisecond * 20, blkHash, types.ResultStatus_OK, 1, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmocks.NewMockActorService(ctrl)
			//mockActor.EXPECT().SendRequest(message.P2PSvc, gomock.Any())
			if test.sentResp > 0 {
				mockActor.EXPECT().TellRequest(message.SyncerSvc, gomock.Any()).DoAndReturn(func(a string, arg *message.GetHashByNoRsp) {
					if !((arg.Err != nil) == test.respError) {
						t.Fatalf("Wrong error (have %v)\n", arg.Err)
					}
				})
			}
			//mockContext.On("Respond",mock.AnythingOfType("*message.GetBlockChunksRsp"))
			mockMF := p2pmocks.NewMockMoFactory(ctrl)
			mockPeer := p2pmocks.NewMockRemotePeer(ctrl)
			//	mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockMo := createDummyMo(ctrl)
			mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(test.consumed)
			mockPeer.EXPECT().SendMessage(gomock.Any())
			mockMF.EXPECT().NewMsgBlockRequestOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)

			//expire := time.Now().Add(test.ttl)
			br := NewBlockHashByNoReceiver(mockActor, mockPeer, test.input, test.ttl)
			br.StartGet()

			msg := &V030Message{subProtocol: subproto.GetHashByNoResponse, id: sampleMsgID}
			body := &types.GetHashByNoResponse{BlockHash: test.blkRsp, Status: test.rspStatus}
			if test.blkInterval > 0 {
				time.Sleep(test.blkInterval)
			}
			br.ReceiveResp(msg, body)

		})
	}
}
