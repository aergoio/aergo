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


func TestSyncWorker_runWorker(t *testing.T) {
	hashes := make([]message.BlockHash, len(sampleTxs))
	for i, hash := range sampleTxs {
		hashes[i] = message.BlockHash(hash)
	}
	stopHash := message.BlockHash(dummyBlockHash)

	tests := []struct {
		name string
		timeout time.Duration
		event func(worker *syncWorker)
	}{
		{"TFinish",  time.Millisecond*300, func(w *syncWorker) {
			time.Sleep(time.Millisecond*10)
			logger.Info().Msg("Sent Event!")
			w.finish<-struct{}{}
		}},
		{"TRetain",  time.Millisecond*100, func(w *syncWorker) {
			time.Sleep(time.Millisecond*10)
			logger.Info().Msg("Sent Event!")
			w.retain<-struct{}{}

		}},
		{"TCancel",  time.Millisecond*100, func(w *syncWorker) {
			time.Sleep(time.Millisecond*10)
			logger.Info().Msg("Sent Event!")
			w.Cancel()

		}},
		{"TTimeout",  time.Millisecond*100, nil},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyMsgID := NewMsgID()
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("TellRequest", message.P2PSvc, mock.AnythingOfType("*message.GetMissingRequest"))
			mockMF := new(MockMoFactory)
			mockMO := new(MockMsgOrder)
			mockMF.On("newMsgRequestOrder",mock.Anything,mock.Anything,mock.Anything).Return(mockMO)
			mockMO.On("GetMsgID").Return(dummyMsgID)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)

			sampleManager := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			target := newSyncWorker(sampleManager, mockPeer,hashes ,stopHash )
			sampleManager.sw = target
			sampleManager.syncing = true

			target.ttl = test.timeout

			go target.runWorker()
			if test.event != nil {
				test.event(target)
			}

			actual, found := sampleManager.getWorker(sampleMeta.ID)
			for found == true {
				time.Sleep(time.Millisecond*50)
				logger.Info().Interface("a",actual).Msg("Waiting finish")
				actual, found = sampleManager.getWorker(sampleMeta.ID)
			}

			assert.Equal(t, dummyMsgID, target.requestMsgID)
			assert.Nil(t, sampleManager.sw)
			assert.False(t, sampleManager.syncing)
		})
	}
}

func TestSyncWorker_putAddBlock(t *testing.T) {
	dummyMsgID := NewMsgID()
	dummyMsg := &V030Message{id:dummyMsgID}
	dummyBlocks := make([]*types.Block, len(sampleTxs))
	hashes := make([]message.BlockHash, len(sampleTxs))

	parentHash := dummyBlockHash
	for i, hash := range sampleTxs {
		hashes[i] = message.BlockHash(hash)
		dummyBlocks[i] = &types.Block{Hash:hash, Header:&types.BlockHeader{PrevBlockHash:parentHash}}
		parentHash = hash
	}
	stopHash := message.BlockHash(dummyBlockHash)
	tests := []struct {
		name string
		parentHash BlkHash
		inputBlocks []*types.Block
		inputNext bool

		tellCnt int
		sentRetain bool
		sentCancel bool
		sentFinish bool
	}{
		// first block add
		{"Tfirst",notDefinedYet, dummyBlocks, false,len(dummyBlocks),true,false,true},
		// success continued blocks
		{"TSucc",MustParseBlkHash(dummyBlockHash), dummyBlocks, false,len(dummyBlocks),true,false,true},
		// success multiple response
		{"TMulti",MustParseBlkHash(dummyBlockHash), dummyBlocks, true,len(dummyBlocks),true,false,false},
		// missing ranges
		{"TMissing",MustParseBlkHash(dummyBlockHash), dummyBlocks[1:], true,0,false,true,false},

		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyRsp := &message.AddBlockRsp{BlockNo:1, Err:nil}
			mockPM := new(MockPeerManager)
			mockActor := new(MockActorService)
			mockActor.On("SendRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"))
			mockActor.On("CallRequest", message.ChainSvc, mock.AnythingOfType("*message.AddBlock"), mock.Anything).Return(dummyRsp, nil)
			mockMF := new(MockMoFactory)
			mockMO := new(MockMsgOrder)
			mockMF.On("newMsgRequestOrder",mock.Anything,mock.Anything,mock.Anything).Return(mockMO)
			mockMO.On("GetMsgID").Return(dummyMsgID)
			mockPeer := new(MockRemotePeer)
			mockPeer.On("Meta").Return(sampleMeta)
			mockPeer.On("ID").Return(sampleMeta.ID)
			mockPeer.On("MF").Return(mockMF)
			mockPeer.On("sendMessage", mock.Anything)

			sampleManager := newSyncManager(mockActor, mockPM, logger).(*syncManager)

			target := newSyncWorker(sampleManager, mockPeer,hashes ,stopHash )
			target.retain = make(chan interface{},10)
			target.cancel = make(chan interface{},10)
			target.finish = make(chan interface{},10)
			sampleManager.sw = target
			sampleManager.syncing = true

			target.currentParent = test.parentHash

			target.putAddBlock(dummyMsg, test.inputBlocks, test.inputNext)

		})
	}
}
