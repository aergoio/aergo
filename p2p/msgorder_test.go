/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_pbRequestOrder_SendTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	factory := &baseMOFactory{}
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

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

			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr)

			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)
			pr := factory.NewMsgRequestOrder(true, p2pcommon.PingRequest, &types.Ping{})
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
	factory := &baseMOFactory{}
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

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
			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr)

			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)
			pr := factory.NewMsgResponseOrder(p2pcommon.NewMsgID(), p2pcommon.PingResponse, &types.Pong{})
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
	factory := &baseMOFactory{}

	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

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
		// no write happen.
		{"TExistWriteFail", fmt.Errorf("writeFail"), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			if tt.keyExist {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(0)
			} else {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(1)
			}
			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)
			peer.lastStatus = &types.LastBlockStatus{}

			target := factory.NewMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: dummyBlockHash, BlockNo: 1})
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

func Test_pbBlkNoticeOrder_SendTo_SkipByHeight(t *testing.T) {
	allSendCnt := 3
	hashes := make([][]byte, allSendCnt)
	for i := 0; i < allSendCnt; i++ {
		token := make([]byte, 32)
		rand.Read(token)
		hashes[i] = token
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	factory := &baseMOFactory{}
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

	tests := []struct {
		name         string
		noDiff       int
		tryCnt       int
		sendInterval time.Duration
		wantSentLow  int // inclusive
		wantSentHigh int // exclusive
		//wantMinSkip int
	}{
		// send all if remote peer is low
		{"TAllLowPeer", -1000, 3, time.Second >> 1, 3, 4},
		//// skip same or higher peer
		//// the first notice is same and skip but seconds will be sent
		{"TSamePeer", 0, 3, time.Second >> 2, 2, 3},
		{"TPartialHigh", 900, 3, time.Second >> 2, 0, 1},
		{"THighPeer", 10000, 3, time.Second >> 2, 0, 1},
		{"TVeryHighPeer", 100000, 3, time.Second >> 2, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			writeCnt := 0
			mockRW.EXPECT().WriteMsg(gomock.Any()).Do(func(arg interface{}) {
				writeCnt++
			}).MinTimes(tt.wantSentLow)

			notiNo := uint64(99999)
			peerBlkNo := uint64(int64(notiNo) + int64(tt.noDiff))
			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)
			peer.lastStatus = &types.LastBlockStatus{BlockNumber: peerBlkNo}

			skipMax := int32(0)
			for i := 0; i < tt.tryCnt; i++ {
				target := factory.NewMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: hashes[i], BlockNo: notiNo + uint64(i)})
				msgID := sampleMsgID
				// notice broadcast is affected by cache
				// put dummy request information in cache
				peer.requests[msgID] = &requestInfo{reqMO: &pbRequestOrder{}}

				if err := target.SendTo(peer); err != nil {
					t.Errorf("pbMessageOrder.SendTo() error = %v, want nil", err)
				}
				if skipMax < peer.skipCnt {
					skipMax = peer.skipCnt
				}
				time.Sleep(tt.sendInterval)
			}
			fmt.Printf("%v : Max skipCnt %v \n", tt.name, skipMax)

			// verification
			if !(tt.wantSentLow <= writeCnt && writeCnt < tt.wantSentHigh) {
				t.Errorf("Sent count %v, want %v:%v", writeCnt, tt.wantSentLow, tt.wantSentHigh)
			}
		})
	}
}

func Test_pbBlkNoticeOrder_SendTo_SkipByTime(t *testing.T) {
	//t.Skip("This test is varied by machine power or load state.")
	allSendCnt := 1000
	hashes := make([][]byte, allSendCnt)
	for i := 0; i < allSendCnt; i++ {
		token := make([]byte, 32)
		rand.Read(token)
		hashes[i] = token
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	factory := &baseMOFactory{}
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

	tests := []struct {
		name     string
		noDiff   int
		tryCnt   int
		wantSent int // inclusive
		//wantMinSkip int
	}{
		{"TLow", -1000, allSendCnt, 4},
		{"TSame", 0, allSendCnt, 4},
		// sent a notice for every 300 times skip
		{"TPartialHigh", 900, allSendCnt, 3},
		// sent a notice for every 3600 times skip
		{"THighAbout1Hour", 4500, allSendCnt, 1},
		// sent a notice for every 3600 times skip, but total try was only 1000, so no send
		{"TVeryHigh", 100000, allSendCnt, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			writeCnt := 0
			mockRW.EXPECT().WriteMsg(gomock.Any()).Do(func(arg interface{}) {
				writeCnt++
			}).Times(tt.wantSent)

			notiNo := uint64(99999)
			peerBlkNo := uint64(int64(notiNo) + int64(tt.noDiff))
			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)
			peer.lastStatus = &types.LastBlockStatus{BlockNumber: peerBlkNo}

			skipMax := int32(0)
			for i := 0; i < tt.tryCnt; i++ {
				target := factory.NewMsgBlkBroadcastOrder(&types.NewBlockNotice{BlockHash: hashes[i], BlockNo: notiNo + uint64(i)})
				msgID := sampleMsgID
				// notice broadcast is affected by cache
				// put dummy request information in cache
				peer.requests[msgID] = &requestInfo{reqMO: &pbRequestOrder{}}

				if err := target.SendTo(peer); err != nil {
					t.Errorf("pbMessageOrder.SendTo() error = %v, want nil", err)
				}
				if skipMax < peer.skipCnt {
					skipMax = peer.skipCnt
				}
				if i&0x0ff == 0 && i > 0 {
					// sleep tree times
					time.Sleep(time.Second >> 2)
				}
			}
			fmt.Printf("%v : Max skipCnt %v \n", tt.name, skipMax)

		})
	}
}

func Test_pbTxNoticeOrder_SendTo(t *testing.T) {
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta := p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}
	sampleConn := p2pcommon.RemoteConn{IP: net.ParseIP("192.168.1.2"), Port: 7846}
	sampleRemote := p2pcommon.RemoteInfo{Meta: sampleMeta, Connection: sampleConn}

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			factory := &baseMOFactory{}
			mockActorServ := p2pmock.NewMockActorService(ctrl)
			mockPeerManager := p2pmock.NewMockPeerManager(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			if tt.keyExist == len(sampleHashes) {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(0)
			} else {
				mockRW.EXPECT().WriteMsg(gomock.Any()).Return(tt.writeErr).Times(1)
			}

			peer := newRemotePeer(sampleRemote, 0, mockPeerManager, mockActorServ, logger, factory, &dummySigner{}, mockRW)

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
