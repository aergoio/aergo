/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlocksChunkReceiver_StartGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inputHashes := make([]message.BlockHash, len(sampleBlks))
	for i, hash := range sampleBlks {
		inputHashes[i] = hash
	}
	tests := []struct {
		name  string
		input []message.BlockHash
		ttl   time.Duration
	}{
		{"TSimple", inputHashes, time.Millisecond * 10},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

			expire := time.Now().Add(test.ttl)
			br := NewBlockReceiver(mockActor, mockPeer, 0, test.input, test.ttl)

			br.StartGet()

			assert.Equal(t, len(test.input), len(br.blockHashes))
			assert.False(t, expire.After(br.timeout))
		})
	}
}

func TestBlocksChunkReceiver_ReceiveResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	chain.Init(1<<20, "", false, 1, 1)

	seqNo := uint64(8723)
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
		name     string
		input    []message.BlockHash
		blkInput [][]*types.Block

		// to verify
		consumed  int
		respError bool
	}{
		{"TSingleResp", inputHashes, [][]*types.Block{inputBlocks}, 1, false},
		{"TMultiResp", inputHashes, [][]*types.Block{inputBlocks[:1], inputBlocks[1:3], inputBlocks[3:]}, 1, false},
		// Fail1 remote err
		{"TRemoteFail", inputHashes, [][]*types.Block{inputBlocks[:0]}, 1, true},
		// server didn't sent last parts. and it is very similar to timeout
		//{"TNotComplete", inputHashes, time.Minute,0,[][]*types.Block{inputBlocks[:2]},1,0, false},
		// Fail2 missing some blocks in the middle
		{"TMissingBlk", inputHashes, [][]*types.Block{inputBlocks[:1], inputBlocks[2:3], inputBlocks[3:]}, 0, true},
		// Fail2-1 missing some blocks in last
		{"TMissingBlkLast", inputHashes, [][]*types.Block{inputBlocks[:1], inputBlocks[1:2], inputBlocks[3:]}, 1, true},
		// Fail3 unexpected block
		{"TDupBlock", inputHashes, [][]*types.Block{inputBlocks[:2], inputBlocks[1:3], inputBlocks[3:]}, 0, true},
		{"TTooManyBlks", inputHashes[:4], [][]*types.Block{inputBlocks[:1], inputBlocks[1:3], inputBlocks[3:]}, 1, true},
		{"TTooManyBlksMiddle", inputHashes[:2], [][]*types.Block{inputBlocks[:1], inputBlocks[1:3], inputBlocks[3:]}, 0, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockActor.EXPECT().TellRequest(message.SyncerSvc, gomock.Any()).
				DoAndReturn(func(a string, arg *message.GetBlockChunksRsp) {
					if !((arg.Err != nil) == test.respError) {
						t.Fatalf("Wrong error (have %v)\n", arg.Err)
					}
					if arg.Seq != seqNo {
						t.Fatalf("Wrong seqNo %d, want %d)\n", arg.Seq, seqNo)
					}
				}).Times(1)

			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMo := createDummyMo(ctrl)
			mockMF.EXPECT().NewMsgRequestOrderWithReceiver(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockMo)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(dummyPeerID).AnyTimes()
			mockPeer.EXPECT().MF().Return(mockMF)
			mockPeer.EXPECT().SendMessage(gomock.Any()).Times(1)
			if test.consumed > 0 {
				mockPeer.EXPECT().ConsumeRequest(gomock.Any()).Times(test.consumed)
			}

			//expire := time.Now().Add(test.ttl)
			br := NewBlockReceiver(mockActor, mockPeer, seqNo, test.input, time.Minute>>1)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetBlocksResponse, sampleMsgID)
			for i, blks := range test.blkInput {
				body := &types.GetBlockResponse{Blocks: blks, HasNext: i < len(test.blkInput)-1}
				br.ReceiveResp(msg, body)
				if br.status == receiverStatusFinished {
					break
				}
			}

		})
	}
}

func TestBlocksChunkReceiver_ReceiveRespTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	chain.Init(1<<20, "", false, 1, 1)

	seqNo := uint64(8723)
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
		name        string
		input       []message.BlockHash
		ttl         time.Duration
		blkInterval time.Duration
		blkInput    [][]*types.Block

		// to verify
		consumed int
		sentResp int
	}{
		// Fail4 response sent after timeout
		{"TBefore", inputHashes, time.Millisecond * 40, time.Millisecond * 100, [][]*types.Block{inputBlocks[:1], inputBlocks[1:3], inputBlocks[3:]}, 1, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//mockContext := new(mockContext)
			mockActor := p2pmock.NewMockActorService(ctrl)
			if test.sentResp > 0 {
				mockActor.EXPECT().TellRequest(message.SyncerSvc, gomock.Any()).
					DoAndReturn(func(a string, arg *message.GetBlockChunksRsp) {
						// timeout should resp with timeout error or not send response
						if arg.Err == nil {
							t.Fatalf("Wrong error (have %v)\n", arg.Err)
						}
						if arg.Seq != seqNo {
							t.Fatalf("Wrong seqNo %d, want %d)\n", arg.Seq, seqNo)
						}
					}).Times(test.sentResp)
			}

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

			//expire := time.Now().Add(test.ttl)
			br := NewBlockReceiver(mockActor, mockPeer, seqNo, test.input, test.ttl)
			br.StartGet()

			msg := p2pcommon.NewSimpleMsgVal(p2pcommon.GetBlocksResponse, sampleMsgID)
			for i, blks := range test.blkInput {
				time.Sleep(test.blkInterval)

				body := &types.GetBlockResponse{Blocks: blks, HasNext: i < len(test.blkInput)-1}
				br.ReceiveResp(msg, body)
				if br.status == receiverStatusFinished {
					break
				}
			}

		})
	}
}
