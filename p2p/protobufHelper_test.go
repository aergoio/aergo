/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_pbRequestOrder_SendTo(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &pbMOFactory{signer: &dummySigner{}}

	tests := []struct {
		name     string
		writeErr error
		want     bool
	}{
		// new request fill cache of peer
		{"TSucc", nil, true},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)

			pr := factory.newMsgRequestOrder(true, PingRequest, &types.Ping{})
			prevCacheSize := len(peer.requests)
			msgID := pr.GetMsgID()

			if got := pr.SendTo(peer); got != tt.want {
				t.Errorf("pbRequestOrder.SendTo() = %v, want %v", got, tt.want)
			}
			if tt.want {
				assert.Equal(t, prevCacheSize+1, len(peer.requests))
				actualMo, ok := peer.requests[msgID]
				assert.True(t, ok)
				assert.Equal(t, pr, actualMo)
			} else {
				assert.Equal(t, prevCacheSize, len(peer.requests))
			}
		})
	}
}

func Test_pbMessageOrder_SendTo(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &pbMOFactory{signer: &dummySigner{}}

	tests := []struct {
		name     string
		writeErr error
		want     bool
	}{
		{"TSucc", nil, true},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)

			pr := factory.newMsgResponseOrder("id"+tt.name, PingResponse, &types.Pong{})
			msgID := pr.GetMsgID()
			// put dummy request information in cache
			peer.requests[msgID] = &pbRequestOrder{}
			prevCacheSize := len(peer.requests)

			if got := pr.SendTo(peer); got != tt.want {
				t.Errorf("pbMessageOrder.SendTo() = %v, want %v", got, tt.want)
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}

func Test_pbBlkNoticeOrder_SendTo(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &pbMOFactory{signer: &dummySigner{}}

	tests := []struct {
		name     string
		writeErr error
		keyExist bool
		want     bool
	}{
		{"TSucc", nil, false, true},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), false, false},
		{"TExist", nil, true, false},
		{"TExistWriteFail", fmt.Errorf("writeFail"), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)
			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)

			target := factory.newMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: dummyBlockHash})
			msgID := sampleMsgID
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &pbRequestOrder{}
			prevCacheSize := len(peer.requests)
			if tt.keyExist {
				var hashKey BlockHash
				copy(hashKey[:], dummyBlockHash)
				peer.blkHashCache.Add(hashKey, true)
			}
			if got := target.SendTo(peer); got != tt.want {
				t.Errorf("pbMessageOrder.SendTo() = %v, want %v", got, tt.want)
			}
			if tt.keyExist {
				mockRW.AssertNotCalled(t, "WriteMsg", mock.Anything)
			} else {
				mockRW.AssertCalled(t, "WriteMsg", mock.AnythingOfType("*types.P2PMessage"))

			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}

func Test_pbTxNoticeOrder_SendTo(t *testing.T) {
	sampleMeta := PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &pbMOFactory{signer: &dummySigner{}}

	sampleHashes := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		sampleHashes[i] = []byte(fmt.Sprint("tx_000", i))
	}
	tests := []struct {
		name     string
		writeErr error
		keyExist int
		want     bool
	}{
		{"TSucc", nil, 0, true},
		{"TWriteFail", fmt.Errorf("writeFail"), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := new(MockActorService)
			mockPeerManager := new(MockPeerManager)

			mockRW := new(MockMsgReadWriter)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeErr)
			peer := newRemotePeer(sampleMeta, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)

			pr := factory.newMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: sampleHashes})
			msgID := pr.GetMsgID()
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &pbRequestOrder{}
			prevCacheSize := len(peer.requests)
			var hashKey [txhashLen]byte
			for i := 0; i < tt.keyExist; i++ {
				copy(hashKey[:], sampleHashes[i])
				peer.txHashCache.Add(hashKey, true)
			}

			if got := pr.SendTo(peer); got != tt.want {
				t.Errorf("pbMessageOrder.SendTo() = %v, want %v", got, tt.want)
			}
			if tt.keyExist == len(sampleHashes) {
				mockRW.AssertNotCalled(t, "WriteMsg", mock.Anything)
			} else {
				mockRW.AssertCalled(t, "WriteMsg", mock.AnythingOfType("*types.P2PMessage"))
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}
