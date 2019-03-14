/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"testing"

	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmocks"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/golang/mock/gomock"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func Test_pbRequestOrder_SendTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &v030MOFactory{}

	tests := []struct {
		name     string
		writeErr error
		wantErr  bool
	}{
		// new request fill cache of peer
		{"TSucc", nil, false},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := p2pmocks.NewMockActorService(ctrl)
			mockPeerManager := p2pmocks.NewMockPeerManager(ctrl)
			mockRW := p2pmocks.NewMockMsgReadWriter(ctrl)
			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr)

			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)
			pr := factory.NewMsgRequestOrder(true, subproto.PingRequest, &types.Ping{})
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, IPAddress: "192.168.1.2", Port: 7845}
	factory := &v030MOFactory{}

	tests := []struct {
		name     string
		writeErr error
		wantErr  bool
	}{
		{"TSucc", nil, false},
		// when failed in send
		{"TWriteFail", fmt.Errorf("writeFail"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := p2pmocks.NewMockActorService(ctrl)
			mockPeerManager := p2pmocks.NewMockPeerManager(ctrl)
			mockRW := p2pmocks.NewMockMsgReadWriter(ctrl)

			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr)

			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)
			pr := factory.NewMsgResponseOrder(p2pcommon.NewMsgID(), subproto.PingResponse, &types.Pong{})
			msgID := pr.GetMsgID()
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO: &pbRequestOrder{}}
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
			mockActorServ := p2pmocks.NewMockActorService(ctrl)
			mockPeerManager := p2pmocks.NewMockPeerManager(ctrl)
			mockRW := p2pmocks.NewMockMsgReadWriter(ctrl)

			if tt.keyExist {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(0)
			} else {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(1)
			}
			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			target := factory.NewMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: dummyBlockHash})
			msgID := sampleMsgID
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO: &pbRequestOrder{}}
			prevCacheSize := len(peer.requests)
			if tt.keyExist {
				hashKey := types.ToBlockID(dummyBlockHash)
				peer.blkHashCache.Add(hashKey, true)
			}
			if err := target.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbMessageOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}

func Test_pbTxNoticeOrder_SendTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
			mockActorServ := p2pmocks.NewMockActorService(ctrl)
			mockPeerManager := p2pmocks.NewMockPeerManager(ctrl)
			mockRW := p2pmocks.NewMockMsgReadWriter(ctrl)

			if tt.keyExist == len(sampleHashes) {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(0)
			} else {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(1)
			}

			peer := newRemotePeer(sampleMeta, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, nil, mockRW)

			pr := factory.NewMsgTxBroadcastOrder(&types.NewTransactionsNotice{TxHashes: sampleHashes})
			msgID := pr.GetMsgID()
			// notice broadcast is affected by cache
			// put dummy request information in cache
			peer.requests[msgID] = &requestInfo{reqMO: &pbRequestOrder{}}
			prevCacheSize := len(peer.requests)
			var hashKey types.TxID
			for i := 0; i < tt.keyExist; i++ {
				hashKey = types.ToTxID(sampleHashes[i])
				peer.txHashCache.Add(hashKey, true)
			}

			if err := pr.SendTo(peer); (err != nil) != tt.wantErr {
				t.Errorf("pbRequestOrder.SendTo() error = %v, wantErr %v", err, tt.wantErr)
			}
			// not affect any cache
			assert.Equal(t, prevCacheSize, len(peer.requests))
			_, ok := peer.requests[msgID]
			assert.True(t, ok)
		})
	}
}
