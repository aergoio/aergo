/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"github.com/aergoio/aergo/p2p/list"
	"github.com/aergoio/aergo/p2p/p2putil"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/network"
)

const (
	OneDay = time.Hour * 24
)

func Test_staticWPManager_OnDiscoveredPeers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		metas []p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TSingleDesign", args{desigPeers[:1]}, 0},
		{"TAllDesign", args{desigPeers}, 0},
		{"TNewID", args{unknownPeers}, 0},
		{"TMixedIDs", args{append(unknownPeers[:5], desigPeers[:5]...)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockLM := p2pmock.NewMockListManager(ctrl)

			dp := NewWaitingPeerManager(logger, dummyPM, mockLM, 10, false).(*staticWPManager)

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_dynamicWPManager_OnDiscoveredPeers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		preConnected []types.PeerID
		metas        []p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TAllNew", args{nil, desigPeers[:1]}, 1},
		{"TAllExist", args{desigIDs, desigPeers[:5]}, 0},
		{"TMixedIDs", args{desigIDs, append(unknownPeers[:5], desigPeers[:5]...)}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockLM := p2pmock.NewMockListManager(ctrl)

			dp := NewWaitingPeerManager(logger, dummyPM, mockLM, 10, true)
			for _, id := range tt.args.preConnected {
				dummyPM.remotePeers[id] = &remotePeerImpl{}
				dp.OnPeerConnect(id)
			}

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_setNextTrial(t *testing.T) {
	dummyDesignated := p2pcommon.PeerMeta{Designated: true}

	type args struct {
		wp     *p2pcommon.WaitingPeer
		setCnt int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TDesig1", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 1}, true},
		{"TDesigSome", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 5}, true},
		{"TDesigMany", args{&p2pcommon.WaitingPeer{Meta: dummyDesignated}, 30}, true},

		{"TUnknown1", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 1}, false},
		{"TUnknownSome", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 5}, false},
		{"TUnknownMany", args{&p2pcommon.WaitingPeer{Meta: dummyMeta}, 30}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastResult := false
			prevDuration := time.Duration(0)
			for i := 0; i < tt.args.setCnt; i++ {
				now := time.Now()
				lastResult = setNextTrial(tt.args.wp)
				gotDuration := tt.args.wp.NextTrial.Sub(now)
				// nextTrial time will be increased exponentially and clipped when trial count is bigger than internal count
				// the clipped
				if lastResult &&
					(gotDuration < prevDuration && gotDuration < OneDay) {
					t.Errorf("smaller duration %v, want at least %v", gotDuration, prevDuration)
				}
				prevDuration = gotDuration
			}

			if lastResult != tt.want {
				t.Errorf("setNextTrial() = %v, want %v", lastResult, tt.want)
			}
		})
	}
}

func Test_basePeerManager_tryAddPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// tests for add peer
	type args struct {
		outbound bool
		meta     p2pcommon.PeerMeta
	}

	tests := []struct {
		name string
		args args

		hsRet *types.Status
		hsErr error

		wantDesign bool
		wantHidden bool
		wantID     types.PeerID
		wantSucc   bool
	}{
		// add inbound peer
		{"TIn", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, false), nil, false, false, dummyPeerID, true},

		// add outbound peer
		{"TOut", args{true, p2pcommon.PeerMeta{ID: dummyPeerID}},
			dummyStatus(dummyPeerID, false), nil, false, false, dummyPeerID, true},

		// failed to handshake
		{"TErrHandshake", args{false, p2pcommon.PeerMeta{ID: dummyPeerID}},
			nil, errors.New("handshake err"), false, false, dummyPeerID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := p2pmock.NewMockStream(ctrl)
			mockHSFactory := p2pmock.NewMockHSHandlerFactory(ctrl)
			mockHSHandler := p2pmock.NewMockHSHandler(ctrl)
			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)
			//mockHSFactory.EXPECT().CreateHSHandler(gomock.Any(), tt.args.outbound, tt.args.meta.ID).Return(mockHSHandler)
			mockHSHandler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(mockRW, tt.hsRet, tt.hsErr)
			//mockHandlerFactory := p2pmock.NewMockHSHandlerFactory(ctrl)
			//mockHandlerFactory.EXPECT().InsertHandlers(gomock.AssignableToTypeOf(&remotePeerImpl{})).MaxTimes(1)

			// in cases of handshake error
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgRequestOrder(false, p2pcommon.GoAway, gomock.Any()).Return(&pbRequestOrder{}).MaxTimes(1)
			mockRW.EXPECT().WriteMsg(gomock.Any()).MaxTimes(1)

			pm := &peerManager{
				hsFactory:      mockHSFactory,
				peerHandshaked: make(chan handshakeResult, 10),
			}
			dpm := &basePeerManager{
				pm:     pm,
				logger: logger,
			}
			got, got1 := dpm.tryAddPeer(tt.args.outbound, tt.args.meta, mockStream, mockHSHandler)
			if got1 != tt.wantSucc {
				t.Errorf("basePeerManager.tryAddPeer() got1 = %v, want %v", got1, tt.wantSucc)
			}
			if tt.wantSucc {
				if got.ID != tt.wantID {
					t.Errorf("basePeerManager.tryAddPeer() got ID = %v, want %v", got.ID, tt.wantID)
				}
				if got.Outbound != tt.args.outbound {
					t.Errorf("basePeerManager.tryAddPeer() got bound = %v, want %v", got.Outbound, tt.args.outbound)
				}
				if got.Designated != tt.wantDesign {
					t.Errorf("basePeerManager.tryAddPeer() got Designated = %v, want %v", got.Designated, tt.wantDesign)
				}
				if got.Hidden != tt.wantHidden {
					t.Errorf("basePeerManager.tryAddPeer() got Hidden = %v, want %v", got.Hidden, tt.wantHidden)
				}
			}

		})
	}
}

func dummyStatus(id types.PeerID, noexpose bool) *types.Status {
	return &types.Status{Sender: &types.PeerAddress{PeerID: []byte(id)}, NoExpose: noexpose}
}

func Test_basePeerManager_CheckAndConnect(t *testing.T) {
	type fields struct {
		pm          *peerManager
		lm          p2pcommon.ListManager
		logger      *log.Logger
		workingJobs map[types.PeerID]ConnWork
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dpm := &basePeerManager{
				pm:          tt.fields.pm,
				lm:          tt.fields.lm,
				logger:      tt.fields.logger,
				workingJobs: tt.fields.workingJobs,
			}
			dpm.CheckAndConnect()
		})
	}
}

func Test_dynamicWPManager_CheckAndConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		preConnected []types.PeerID
		metas        []p2pcommon.PeerMeta
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{"TAllNew", args{nil, desigPeers[:1]}, 1},
		{"TAllExist", args{desigIDs, desigPeers[:5]}, 0},
		{"TMixedIDs", args{desigIDs, append(unknownPeers[:5], desigPeers[:5]...)}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := createDummyPM()
			mockLM := p2pmock.NewMockListManager(ctrl)

			dp := NewWaitingPeerManager(logger, dummyPM, mockLM, 10, true)
			for _, id := range tt.args.preConnected {
				dummyPM.remotePeers[id] = &remotePeerImpl{}
				dp.OnPeerConnect(id)
			}

			dp.OnDiscoveredPeers(tt.args.metas)
			if len(dummyPM.waitingPeers) != tt.wantCount {
				t.Errorf("count waitingPeer %v, want %v", len(dummyPM.waitingPeers), tt.wantCount)
			}
		})
	}
}

func Test_basePeerManager_connectWaitingPeers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := []*p2pcommon.WaitingPeer{}
	for i := 0 ; i < 5 ; i++ {
		meta := p2pcommon.PeerMeta{ID:p2putil.RandomPeerID()}
		wp := &p2pcommon.WaitingPeer{Meta:meta, NextTrial:time.Now()}
		c = append(c, wp)
	}
	n := []*p2pcommon.WaitingPeer{}
	for i := 0 ; i < 5 ; i++ {
		meta := p2pcommon.PeerMeta{ID:p2putil.RandomPeerID()}
		wp := &p2pcommon.WaitingPeer{Meta:meta, NextTrial:time.Now().Add(time.Hour*100)}
		n = append(n, wp)
	}
	type args struct {
		maxJob int
	}
	tests := []struct {
		name string
		wjs  []*p2pcommon.WaitingPeer
		args args

		wantCnt int
	}{
		{"TEmptyJob", nil, args{4}, 0},
		{"TFewer", c[:2], args{4}, 2},
		{"TLarger", c, args{4}, 4},
		{"TWithNotConn", append(nc(),c[0],n[0],n[1],c[1],n[4],c[4]), args{4}, 3},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLM := p2pmock.NewMockListManager(ctrl)
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			wpMap := make(map[types.PeerID]*p2pcommon.WaitingPeer)
			for _, w := range tt.wjs {
				wpMap[w.Meta.ID] = w
			}

			dummyPM := &peerManager{nt: mockNT, waitingPeers:wpMap, workDoneChannel: make(chan p2pcommon.ConnWorkResult,10)}

			dpm := &basePeerManager{
				pm:          dummyPM,
				lm:          mockLM,
				logger:      logger,
				workingJobs: make(map[types.PeerID]ConnWork),
			}

			mockNT.EXPECT().GetOrCreateStream(gomock.Any(), gomock.Any()).Return(nil, errors.New("stream failed")).Times(tt.wantCnt)
			mockLM.EXPECT().IsBanned(gomock.Any(), gomock.Any()).Return(false, list.FarawayFuture).AnyTimes()

			dpm.connectWaitingPeers(tt.args.maxJob)

			doneCnt := 0
			expire := time.NewTimer(time.Millisecond * 500)
			WAITLOOP:
			for doneCnt < tt.wantCnt {
				select {
				case <-dummyPM.workDoneChannel:
					doneCnt++
				case <-expire.C:
					t.Errorf("connectWaitingPeers() job cnt %v, want %v", doneCnt, tt.wantCnt)
					break WAITLOOP
				}
			}
		})
	}
}

func nc() []*p2pcommon.WaitingPeer {
	return nil
}
func Test_basePeerManager_OnInboundConn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		workingJobs map[types.PeerID]ConnWork
	}
	type args struct {
		s network.Stream
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyPM := &peerManager{}
			mockLM := p2pmock.NewMockListManager(ctrl)

			mockStream := p2pmock.NewMockStream(ctrl)

			dpm := &basePeerManager{
				pm:          dummyPM,
				lm:          mockLM,
				logger:      logger,
				workingJobs: tt.fields.workingJobs,
			}
			dpm.OnInboundConn(mockStream)
		})
	}
}
