/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	network2 "github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/list"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/network"
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
			mockIS := p2pmock.NewMockInternalService(ctrl)

			dp := NewWaitingPeerManager(logger, mockIS, dummyPM, mockLM, 10, false).(*staticWPManager)

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
			mockIS := p2pmock.NewMockInternalService(ctrl)

			dp := NewWaitingPeerManager(logger, mockIS, dummyPM, mockLM, 10, true)
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
	type args struct {
		wp     *p2pcommon.WaitingPeer
		setCnt int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TDesig1", args{&p2pcommon.WaitingPeer{Meta: dummyMeta, Designated: true}, 1}, true},
		{"TDesigSome", args{&p2pcommon.WaitingPeer{Meta: dummyMeta, Designated: true}, 5}, true},
		{"TDesigMany", args{&p2pcommon.WaitingPeer{Meta: dummyMeta, Designated: true}, 30}, true},

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

		hsRet *p2pcommon.HandshakeResult
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
			mockHSHandler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(tt.hsRet, tt.hsErr)
			//mockHandlerFactory := p2pmock.NewMockHSHandlerFactory(ctrl)
			//mockHandlerFactory.EXPECT().InsertHandlers(gomock.AssignableToTypeOf(&remotePeerImpl{})).MaxTimes(1)
			if tt.hsErr == nil {
				mockConn := p2pmock.NewMockConn(ctrl)
				dummyMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
				mockStream.EXPECT().Conn().Return(mockConn)
				mockConn.EXPECT().RemoteMultiaddr().Return(dummyMA)
			}

			// in cases of handshake error
			mockMF := p2pmock.NewMockMoFactory(ctrl)
			mockMF.EXPECT().NewMsgRequestOrder(false, p2pcommon.GoAway, gomock.Any()).Return(&pbRequestOrder{}).MaxTimes(1)
			mockRW.EXPECT().WriteMsg(gomock.Any()).MaxTimes(1)
			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockIS.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{}).AnyTimes()

			pm := &peerManager{
				hsFactory:     mockHSFactory,
				peerConnected: make(chan connPeerResult, 10),
			}
			dpm := &basePeerManager{
				pm:     pm,
				is:     mockIS,
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
				if got.Hidden != tt.wantHidden {
					t.Errorf("basePeerManager.tryAddPeer() got Hidden = %v, want %v", got.Hidden, tt.wantHidden)
				}
			}

		})
	}
}

func dummyStatus(id types.PeerID, noexpose bool) *p2pcommon.HandshakeResult {
	return &p2pcommon.HandshakeResult{Meta: p2pcommon.PeerMeta{ID: id}, Hidden: noexpose}
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
			mockIS := p2pmock.NewMockInternalService(ctrl)

			dp := NewWaitingPeerManager(logger, mockIS, dummyPM, mockLM, 10, true)
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
	for i := 0; i < 5; i++ {
		meta := p2pcommon.PeerMeta{ID: types.RandomPeerID()}
		wp := &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now()}
		c = append(c, wp)
	}
	n := []*p2pcommon.WaitingPeer{}
	for i := 0; i < 5; i++ {
		meta := p2pcommon.PeerMeta{ID: types.RandomPeerID()}
		wp := &p2pcommon.WaitingPeer{Meta: meta, NextTrial: time.Now().Add(time.Hour * 100)}
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
		{"TWithNotConn", append(nc(), c[0], n[0], n[1], c[1], n[4], c[4]), args{4}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLM := p2pmock.NewMockListManager(ctrl)
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			wpMap := make(map[types.PeerID]*p2pcommon.WaitingPeer)
			for _, w := range tt.wjs {
				wpMap[w.Meta.ID] = w
			}

			dummyPM := &peerManager{nt: mockNT, waitingPeers: wpMap, workDoneChannel: make(chan p2pcommon.ConnWorkResult, 10)}

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

func Test_basePeerManager_createRemoteInfo(t *testing.T) {
	i4, n4, _ := net.ParseCIDR("192.56.1.1/24")
	innerIP := i4.String()

	pid1 := types.RandomPeerID()
	type args struct {
		status   p2pcommon.HandshakeResult
		outbound bool
	}
	tests := []struct {
		name       string
		args       args
		ip         string
		port       uint32
		wantHidden bool
	}{
		{"TOut1", args{p2pcommon.HandshakeResult{Meta: p2pcommon.PeerMeta{ID: pid1}, Hidden: false}, true},
			innerIP, 7846, false},
		{"TOutHidden", args{p2pcommon.HandshakeResult{Meta: p2pcommon.PeerMeta{ID: pid1}, Hidden: true}, true}, innerIP, 7846, true},
		{"TIn", args{p2pcommon.HandshakeResult{Meta: p2pcommon.PeerMeta{ID: pid1}, Hidden: false}, false},
			innerIP, 7846, false},
		{"TInHidden", args{p2pcommon.HandshakeResult{Meta: p2pcommon.PeerMeta{ID: pid1}, Hidden: true}, true}, innerIP, 7846, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conn := p2pmock.NewMockConn(ctrl)
			ma, _ := types.ToMultiAddr(tt.ip, tt.port)
			conn.EXPECT().RemoteMultiaddr().Return(ma).AnyTimes()
			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockIS.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{InternalZones: []*net.IPNet{n4}}).AnyTimes()

			dpm := &basePeerManager{
				logger: logger,
				is:     mockIS,
			}
			got := dpm.createRemoteInfo(conn, tt.args.status, tt.args.outbound)
			if !types.IsSamePeerID(got.Meta.ID, types.PeerID(pid1)) {
				t.Errorf("createRemoteInfo() ID = %v, want %v", got.Meta.ID.String(), pid1)
			}
			if got.Hidden != tt.wantHidden {
				t.Errorf("createRemoteInfo() hidden = %v, want %v", got.Hidden, tt.wantHidden)
			}
			if got.Connection.Outbound != tt.args.outbound {
				t.Errorf("createRemoteInfo() outbound = %v, want %v", got.Connection.Outbound, tt.args.outbound)
			}
			if !network2.IsSameAddress(got.Connection.IP.String(), tt.ip) {
				t.Errorf("createRemoteInfo() addr = %v, want %v", got.Connection.IP, tt.ip)
			}
		})
	}
}

func Test_basePeerManager_createRemoteInfoOfZone(t *testing.T) {
	i4, n4, _ := net.ParseCIDR("192.56.1.1/24")
	innerIP := i4.String()
	innerIP2 := "192.56.1.200"
	externalIP1 := "192.56.2.1"

	pid1 := types.RandomPeerID()
	sampleMeta := p2pcommon.PeerMeta{ID: pid1}
	type args struct {
		status p2pcommon.HandshakeResult
	}
	tests := []struct {
		name     string
		args     args
		ip       string
		wantZone p2pcommon.PeerZone
	}{
		{"TInternal1", args{p2pcommon.HandshakeResult{Meta: sampleMeta, Hidden: false}},
			innerIP, p2pcommon.InternalZone},
		{"TInternal2", args{p2pcommon.HandshakeResult{Meta: sampleMeta, Hidden: false}},
			innerIP2, p2pcommon.InternalZone},
		{"TExternal", args{p2pcommon.HandshakeResult{Meta: sampleMeta, Hidden: false}},
			externalIP1, p2pcommon.ExternalZone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			conn := p2pmock.NewMockConn(ctrl)
			ma, _ := types.ToMultiAddr(tt.ip, 7846)
			conn.EXPECT().RemoteMultiaddr().Return(ma).AnyTimes()
			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockIS.EXPECT().LocalSettings().Return(p2pcommon.LocalSettings{InternalZones: []*net.IPNet{n4}}).AnyTimes()

			dpm := &basePeerManager{
				logger: logger,
				is:     mockIS,
			}
			got := dpm.createRemoteInfo(conn, tt.args.status, false)
			if !types.IsSamePeerID(got.Meta.ID, types.PeerID(pid1)) {
				t.Errorf("createRemoteInfo() ID = %v, want %v", got.Meta.ID.String(), pid1)
			}
			if got.Zone != tt.wantZone {
				t.Errorf("createRemoteInfo() Zone = %v, want %v", got.Zone, tt.wantZone)
			}
		})
	}
}

func Test_basePeerManager_OnWorkDone(t *testing.T) {
	logger := log.NewLogger("p2p.test")
	pid := types.RandomPeerID()
	meta := p2pcommon.NewMetaWith1Addr(pid, "192.168.2.3", 7846, "v2.0.0")
	wpTmpl := p2pcommon.WaitingPeer{Meta: meta}
	argTmpl := p2pcommon.ConnWorkResult{Meta: meta}
	type args struct {
		designated bool
		inbound    bool
		succeed    bool
	}
	tests := []struct {
		name      string
		args      args
		wantRetry bool
	}{
		// 1. work success
		{"TInSucc", args{false, true, true}, false},
		{"TOutSucc", args{false, true, true}, false},
		// 2. work failed of designated peer
		{"TInFailDesignated", args{true, true, false}, true},
		{"TOutFailDesignated", args{true, false, false}, true},
		// 3. work failed of normal peer
		{"TInFailNormal", args{false, true, false}, false},
		{"TOutFailNormal", args{false, false, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			dummyPM := createDummyPM()
			mockLM := p2pmock.NewMockListManager(ctrl)
			mockIS := p2pmock.NewMockInternalService(ctrl)

			dpm := &basePeerManager{
				is:          mockIS,
				pm:          dummyPM,
				lm:          mockLM,
				logger:      logger,
				workingJobs: make(map[types.PeerID]ConnWork),
			}
			dpm.workingJobs[pid] = ConnWork{Meta: meta, PeerID: pid, StartTime: time.Now().Add(-time.Millisecond)}
			wp := wpTmpl
			wp.Designated = tt.args.designated
			dummyPM.waitingPeers[pid] = &wp
			arg := argTmpl
			arg.Inbound = tt.args.inbound
			if !tt.args.succeed {
				arg.Result = sampleErr
			}

			dpm.OnWorkDone(arg)
			if _, exist := dpm.workingJobs[pid]; exist {
				t.Errorf("OnWorkDone() job must be deleted, but not")
			}
			if _, exist := dummyPM.waitingPeers[pid]; exist != tt.wantRetry {
				t.Errorf("OnWorkDone() wantRetry %v , but not", tt.wantRetry)
			}

		})
	}
}

func Test_dynamicWPManager_OnPeerDisconnect(t *testing.T) {
	pid1, pid2 := types.RandomPeerID(), types.RandomPeerID()
	meta1 := p2pcommon.NewMetaWith1Addr(pid1, "127.0.0.1", 7846, "v2.0")
	meta2 := p2pcommon.NewMetaWith1Addr(pid2, "127.0.0.2", 7846, "v2.0")

	type args struct {
		meta p2pcommon.PeerMeta
	}
	tests := []struct {
		name string
		args args

		wantRetry bool
	}{
		{"TDesignated", args{meta1}, true},
		{"TDesignated", args{meta2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			dummyPM := createDummyPM()
			dummyPM.designatedPeers = make(map[types.PeerID]p2pcommon.PeerMeta)
			dummyPM.designatedPeers[pid1] = meta1

			mockLM := p2pmock.NewMockListManager(ctrl)
			mockIS := p2pmock.NewMockInternalService(ctrl)
			mockPeer := p2pmock.NewMockRemotePeer(ctrl)
			mockPeer.EXPECT().ID().Return(tt.args.meta.ID).AnyTimes()
			mockPeer.EXPECT().Meta().Return(tt.args.meta).AnyTimes()
			mockPeer.EXPECT().Name().Return(p2putil.ShortMetaForm(tt.args.meta)).AnyTimes()

			dpm := NewWaitingPeerManager(logger, mockIS, dummyPM, mockLM, 4, true)

			dpm.OnPeerDisconnect(mockPeer)

			if _, exist := dummyPM.waitingPeers[tt.args.meta.ID]; exist != tt.wantRetry {
				t.Errorf("OnPeerDisconnect() wantRetry %v , but not", tt.wantRetry)
			}
		})
	}
}
