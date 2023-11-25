/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAncestorReceiver_StartGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	seqNo := uint64(777)

	tests := []struct {
		name  string
		input [][]byte
		ttl   time.Duration
	}{
		{"TSimple", sampleBlks, time.Millisecond * 10},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)

			mockMo := createDummyMo(ctrl)
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(mockMo).Times(1)

			expire := time.Now().Add(test.ttl)
			br := NewAncestorReceiver(mockActor, mockPeer, seqNo, test.input, test.ttl)

			br.StartGet()

			assert.False(t, expire.After(br.timeout))

			// getBlock must be sent
		})
	}
}

func TestAncestorReceiver_ReceiveResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	seqNo := uint64(33)
	blkHash := dummyBlockHash

	tests := []struct {
		name        string
		input       [][]byte
		ttl         time.Duration
		blkInterval time.Duration

		blkRsp    []byte
		blkNo     types.BlockNo
		rspStatus types.ResultStatus

		// to verify
		consumed int
		sentResp int
		respNil  bool
	}{
		{"TSame", sampleBlks, time.Minute, 0, blkHash, 12, types.ResultStatus_OK, 1, 1, false},
		// Fail1 remote err
		{"TFirst", sampleBlks, time.Minute, 0, nil, 0, types.ResultStatus_INTERNAL, 1, 1, true},
		// Fail2 can't find block
		{"TNotMatch", sampleBlks, time.Minute, 0, nil, 0, types.ResultStatus_NOT_FOUND, 1, 1, true},
		// Fail4 response sent after timeout
		{"TTimeout", sampleBlks, time.Millisecond * 10, time.Millisecond * 20, blkHash, 12, types.ResultStatus_OK, 1, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			//mockActor.EXPECT().SendRequest(message.P2PSvc, gomock.Any())
			if test.sentResp > 0 {
				mockActor.EXPECT().TellRequest(message.SyncerSvc, gomock.Any()).DoAndReturn(func(a string, arg *message.GetSyncAncestorRsp) {
					if !((arg.Ancestor == nil) == test.respNil) {
						t.Fatalf("Wrong error (have %v)\n", arg.Ancestor)
					}
					if arg.Seq != seqNo {
						t.Fatalf("Wrong seqNo %d, want %d)\n", arg.Seq, seqNo)
					}
				})
			}
			//mockContext.On("Respond",mock.AnythingOfType("*message.GetBlockChunksRsp"))
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			//	mockPeer.On("ID").Return(dummyPeerID)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockMo := createDummyMo(ctrl)
			mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(test.consumed)
			mockPeer.EXPECT().SendMessage(gomock.Any())
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)

			//expire := time.Now().Add(test.ttl)
			br := NewAncestorReceiver(mockActor, mockPeer, seqNo, test.input, test.ttl)
			br.StartGet()

			msg := p2pcommon.NewMessageValue(p2pcommon.GetAncestorResponse, sampleMsgID, p2pcommon.EmptyID, time.Now().UnixNano(), nil)
			body := &types.GetAncestorResponse{AncestorHash: test.blkRsp, AncestorNo: test.blkNo, Status: test.rspStatus}
			if test.blkInterval > 0 {
				time.Sleep(test.blkInterval)
			}
			br.ReceiveResp(msg, body)

		})
	}
}
