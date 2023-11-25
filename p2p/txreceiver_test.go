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

func TestGetTxsReceiver_StartGet(t *testing.T) {

	inputHashes := make([]types.TxID, len(sampleBlks))
	inTXs := make([]*types.Tx, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = types.ToTxID(hash)
		inTXs[i] = &types.Tx{Hash: hash, Body: &types.TxBody{}}
	}
	tests := []struct {
		name  string
		input []types.TxID
		ttl   time.Duration
	}{
		{"TSimple", inputHashes, time.Millisecond * 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			//mockActor.On("SendRequest", message.P2PSvc, mock.AnythingOfType("*types.GetBlock"))
			//mockActor.On("TellRequest", message.SyncerSvc, mock.AnythingOfType("*types.GetBlock"))
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(mockMo).Times(1)
			mockSM := p2pmock.NewMockSyncManager(ctrl)

			expire := time.Now().Add(tt.ttl)
			br := NewGetTxsReceiver(mockActor, mockPeer, mockSM, logger, tt.input, tt.ttl)

			br.StartGet()

			assert.Equal(t, len(tt.input), len(br.hashes))
			assert.False(t, expire.After(br.timeout))
		})
	}
}

func TestGetTxsReceiver_ReceiveResp(t *testing.T) {
	inputHashes := make([]types.TxID, len(sampleTxs))
	inTXs := make([]*types.Tx, len(sampleTxs))
	for i, hash := range sampleTxs {
		inputHashes[i] = types.ToTxID(hash)
		inTXs[i] = &types.Tx{Hash: hash, Body: &types.TxBody{}}
	}
	inSize := len(inTXs)
	tests := []struct {
		name     string
		input    []types.TxID
		blkInput [][]*types.Tx

		// to verify
		putCnt      int
		wantMiss    int
		wantErr     bool
		wantConsume bool
	}{
		{"TSingleResp", inputHashes, [][]*types.Tx{inTXs}, inSize, 0, false, true},
		{"TMultiResp", inputHashes, [][]*types.Tx{inTXs[:1], inTXs[1:3], inTXs[3:]}, inSize, 0, false, true},
		// Fail1 remote err
		{"TRemoteFail", inputHashes, [][]*types.Tx{inTXs[:0]}, 0, 0, true, true},
		// server didn't sent last parts. and it is very similar to timeout
		//{"TNotComplete", inputHashes, time.Minute,0,[][]*types.Block{inTXs[:2]},1,0, false},
		// Fail2 missing some blocks in the middle
		{"TMissingBlk", inputHashes, [][]*types.Tx{inTXs[:1], inTXs[2:3], inTXs[3:]}, inSize - 1, 1, false, true},
		// Fail2-1 missing some blocks in last
		{"TMissingBlkLast", inputHashes, [][]*types.Tx{inTXs[:1], inTXs[1:2], inTXs[3:]}, inSize - 1, 1, false, true},
		// Fail3 unexpected block
		{"TDupBlock", inputHashes, [][]*types.Tx{inTXs[:2], inTXs[1:3], inTXs[3:]}, 2, 0, true, false},
		{"TTooManyBlks", inputHashes[:4], [][]*types.Tx{inTXs[:1], inTXs[1:3], inTXs[3:]}, 4, 0, true, true},
		{"TTooManyBlksMiddle", inputHashes[:2], [][]*types.Tx{inTXs[:1], inTXs[1:3], inTXs[3:]}, 2, 0, true, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

			if test.wantConsume {
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(1)
			}
			mockActor.EXPECT().SendRequest(message.MemPoolSvc, gomock.Any()).Times(test.putCnt)

			mockSM := p2pmock.NewMockSyncManager(ctrl)

			//expire := time.Now().Add(test.ttl)
			br := NewGetTxsReceiver(mockActor, mockPeer, mockSM, logger, test.input, time.Hour>>1)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetTXsResponse, sampleMsgID)
			for i, txs := range test.blkInput {
				hashes := make([][]byte, len(txs))
				for i, b := range txs {
					hashes[i] = b.Hash
				}
				body := &types.GetTransactionsResponse{Hashes: hashes, Txs: txs, HasNext: i < len(test.blkInput)-1}
				br.ReceiveResp(msg, body)
				if br.status == receiverStatusFinished {
					break
				}
			}

			if !test.wantErr && len(br.missed) != test.wantMiss {
				t.Errorf("wantMiss tx cnt %v, wnat %v ", len(br.missed), test.wantMiss)
			}
		})
	}
}

func TestGetTxsReceiver_ReceiveRespBusyRemote(t *testing.T) {
	inputHashes := make([]types.TxID, len(sampleBlks))
	inTXs := make([]*types.Tx, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = types.ToTxID(hash)
		inTXs[i] = &types.Tx{Hash: hash, Body: &types.TxBody{}}
	}
	tests := []struct {
		name   string
		stCode types.ResultStatus

		// to verify
		wantReGet   bool
		wantErr     bool
		wantConsume bool
	}{
		{"TRemoteBusy", types.ResultStatus_RESOURCE_EXHAUSTED, true, true, true},
		{"TNotFound", types.ResultStatus_NOT_FOUND, false, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)

			if test.wantConsume {
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(1)
			}

			mockSM := p2pmock.NewMockSyncManager(ctrl)
			if test.wantReGet {
				mockSM.EXPECT().RetryGetTx(gomock.Any(), gomock.AssignableToTypeOf([][]byte{})).Times(1)
			}
			//expire := time.Now().Add(test.ttl)
			br := NewGetTxsReceiver(mockActor, mockPeer, mockSM, logger, inputHashes, time.Minute>>1)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetTXsResponse, sampleMsgID)
			body := &types.GetTransactionsResponse{Hashes: nil, Txs: nil, HasNext: false, Status: test.stCode}
			br.ReceiveResp(msg, body)

		})
	}
}

func TestGetTxsReceiver_ReceiveRespTimeout(t *testing.T) {
	inputHashes := make([]types.TxID, len(sampleBlks))
	inputTXs := make([]*types.Tx, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = types.ToTxID(hash)
		inputTXs[i] = &types.Tx{Hash: hash, Body: &types.TxBody{}}
	}

	tests := []struct {
		name        string
		input       []types.TxID
		ttl         time.Duration
		blkInterval time.Duration
		blkInput    [][]*types.Tx

		// to verify
		consumed int
		sentResp int
	}{
		// Fail4 response sent after timeout
		{"TBefore", inputHashes, time.Millisecond * 40, time.Millisecond * 100, [][]*types.Tx{inputTXs[:1], inputTXs[1:3], inputTXs[3:]}, 1, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			if test.consumed > 0 {
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).MinTimes(test.consumed)
			}
			mockSM := p2pmock.NewMockSyncManager(ctrl)
			//expire := time.Now().Add(test.ttl)
			br := NewGetTxsReceiver(mockActor, mockPeer, mockSM, logger, test.input, test.ttl)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetTXsResponse, sampleMsgID)
			for i, blks := range test.blkInput {
				time.Sleep(test.blkInterval)

				body := &types.GetTransactionsResponse{Txs: blks, HasNext: i < len(test.blkInput)-1}
				br.ReceiveResp(msg, body)
				if br.status == receiverStatusFinished {
					break
				}
			}

		})
	}
}
