/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_pbRequestOrder_SendTo(t *testing.T) {
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &v030MOFactory{}

	tests := []struct {
		name     string
		writeErr error
		wantErr   bool
	}{
		// new request fill cache of peer
		{"TSucc", nil, false},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			pr := factory.newMsgRequestOrder(true, PingRequest, &types.Ping{})
			prevCacheSize := len(peer.requests)
			msgID := pr.GetMsgID()

			if err := pr.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbRequestOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, prevCacheSize+1, len(peer.requests))
				actualMo, ok := peer.requests[msgID]
				assert.True(t, ok)
				assert.Equal(t, pr, actualMo.reqMO)
			} else {
				assert.Equal(t, prevCacheSize, len(peer.requests))
			}
		})
	}
}

func Test_pbMessageOrder_SendTo(t *testing.T) {
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory :=&v030MOFactory{}

	tests := []struct {
		name     string
		writeErr error
		wantErr     bool
	}{
		{"TSucc", nil, false},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			pr := factory.newMsgResponseOrder(p2pcommon.NewMsgID(), PingResponse, &types.Pong{})
			msgID := pr.GetMsgID()
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO:&pbRequestOrder{}}
			prevCacheSize := len(peer.requests)

			if err := pr.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbMessageOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}

func Test_pbBlkNoticeOrder_SendTo(t *testing.T) {
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &v030MOFactory{}

	tests := []struct {
		name     string
		writeErr error
		keyExist bool
		wantErr  bool
	}{
		{"TSucc", nil, false, false},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), false, true},
		{"TExist", nil, true, false},
		// no write occured.
		{"TExistWriteFail", fmt.Errorf("writeFail"), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)
			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			target := factory.newMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: dummyBlockHash})
			msgID := sampleMsgID
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO:&pbRequestOrder{}}
			prevCacheSize := len(peer.requests)
			if tt.keyExist {
				hashKey := types.ToBlockID(dummyBlockHash)
				peer.blkHashCache.Add(hashKey, true)
			}
			if err := target.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbMessageOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.keyExist {
				mockRW.AssertNotCalled(t, "WriteMsg", mock.Anything)
			} else {
				mockRW.AssertCalled(t, "WriteMsg", mock.AnythingOfType("*p2p.V030Message"))

			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}

func Test_pbTxNoticeOrder_SendTo(t *testing.T) {
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &v030MOFactory{}

	sampleHashes := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		sampleHashes[i] = []byte(fmt.Sprint("tx_000", i))
	}
	tests := []struct {
		name     string
		writeErr error
		keyExist int
		wantErr  bool
	}{
		{"TSucc", nil, 0, false},
		{"TWriteFail", fmt.Errorf("writeFail"), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			pr := factory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: sampleHashes})
			msgID := pr.GetMsgID()
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO:&pbRequestOrder{}}
			prevCacheSize := len(peer.requests)
			var hashKey types.TxID
			for i := 0; i < tt.keyExist; i++ {
				hashKey = types.ToTxID(sampleHashes[i])
				peer.txHashCache.Add(hashKey, true)
			}

			if err := pr.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbRequestOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.keyExist == len(sampleHashes) {
				mockRW.AssertNotCalled(t, "WriteMsg", mock.Anything)
			} else {
				mockRW.AssertCalled(t, "WriteMsg", mock.AnythingOfType("*p2p.V030Message"))
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}
