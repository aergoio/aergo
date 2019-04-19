package p2p

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

func FailTestGetPeers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockActorServ := p2pmock.NewMockActorService(ctrl)
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.EXPECT().CallRequest(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	mockMF := p2pmock.NewMockMoFactory(ctrl)
	target := NewPeerManager(nil, nil, mockActorServ,
		cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config),
		nil, nil, nil,
		log.NewLogger("test.p2p"), mockMF).(*peerManager)

	iterSize := 500
	wg := sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			peerID := peer.ID(strconv.Itoa(i))
			peerMeta := p2pcommon.PeerMeta{ID: peerID}
			target.remotePeers[peerID] = newRemotePeer(peerMeta, 0, target, mockActorServ, logger, nil, nil, nil, nil)
			if i == (iterSize >> 2) {
				wg.Done()
			}
		}
	}()

	go func() {
		wg.Wait()
		for key, val := range target.remotePeers {
			fmt.Printf("%s is %s\n", key.String(), val.State().String())
		}
		waitChan <- 0
	}()

	<-waitChan
}

func TestPeerManager_GetPeers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockActorServ := p2pmock.NewMockActorService(ctrl)
	mockMF := p2pmock.NewMockMoFactory(ctrl)

	tLogger := log.NewLogger("test.p2p")
	tConfig := cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config)
	InitNodeInfo(&tConfig.BaseConfig, tConfig.P2P, tLogger)
	target := NewPeerManager(nil, nil, mockActorServ,
		tConfig,
		nil, nil, nil,
		tLogger, mockMF).(*peerManager)

	iterSize := 500
	wg := &sync.WaitGroup{}
	wgAll := &sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	wgAll.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			peerID := peer.ID(strconv.Itoa(i))
			peerMeta := p2pcommon.PeerMeta{ID: peerID}
			target.insertPeer(peerID, newRemotePeer(peerMeta, 0, target, mockActorServ, logger, nil, nil, nil, nil))
			if i == (iterSize >> 2) {
				wg.Done()
			}
		}
		wgAll.Done()
	}()

	cnt := 0
	go func() {
		wg.Wait()
		for _ = range target.GetPeers() {
			cnt++
		}
		assert.True(t, cnt > (iterSize >> 2))
		waitChan <- 0
	}()

	<-waitChan

	wgAll.Wait()
	assert.True(t, iterSize == len(target.GetPeers()))
}

func TestPeerManager_GetPeerAddresses(t *testing.T) {
	peersLen := 3
	samplePeers := make([]*remotePeerImpl, peersLen)
	samplePeers[0] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID}, lastStatus: &types.LastBlockStatus{}}
	samplePeers[1] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID2}, lastStatus: &types.LastBlockStatus{}}
	samplePeers[2] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID3}, lastStatus: &types.LastBlockStatus{}}
	tests := []struct {
		name string
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pm := &peerManager{remotePeers: make(map[peer.ID]p2pcommon.RemotePeer)}
			for _, peer := range samplePeers {
				pm.remotePeers[peer.ID()] = peer
			}

			actPeers := pm.GetPeerAddresses(false, false)
			assert.Equal(t, peersLen, len(actPeers))
		})
	}
}

func TestPeerManager_init(t *testing.T) {
	tConfig := cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config)
	defaultCfg := tConfig.P2P
	InitNodeInfo(&tConfig.BaseConfig, defaultCfg, logger)
	localIP, _ := p2putil.ExternalIP()

	tests := []struct {
		name            string
		inCfg           *cfg.P2PConfig
		expectProtoAddr string
		expectProtoPort uint32
		expectBindAddr  string
		expectBindPort  uint32
		expectPanic     bool
	}{
		{"TDefault", defaultCfg, localIP.String(), uint32(defaultCfg.NetProtocolPort), localIP.String(), uint32(defaultCfg.NetProtocolPort), false},
		// wrong ProtocolAddress 0.0.0.0
		{"TUnspecifiedAddr", &cfg.P2PConfig{NetProtocolAddr: "0.0.0.0", NetProtocolPort: 7846}, localIP.String(), 7846, localIP.String(), uint32(defaultCfg.NetProtocolPort), true},
		// wrong ProtocolAddress
		{"TWrongAddr", &cfg.P2PConfig{NetProtocolAddr: "24558.30.0.0", NetProtocolPort: 7846}, localIP.String(), 7846, localIP.String(), 7846, true},
		// bind all address
		{"TBindAll", &cfg.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7846, NPBindAddr: "0.0.0.0"}, localIP.String(), 7846, "0.0.0.0", 7846, false},
		// bind differnt address
		{"TBindDifferAddr", &cfg.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7846, NPBindAddr: "172.21.1.2"}, localIP.String(), 7846, "172.21.1.2", 7846, false},
		// bind different port
		{"TDifferPort", &cfg.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7846, NPBindPort: 12345}, localIP.String(), 7846, localIP.String(), 12345, false},
		// bind different address and port
		{"TBindDiffer", &cfg.P2PConfig{NetProtocolAddr: "", NetProtocolPort: 7846, NPBindAddr: "172.21.1.2", NPBindPort: 12345}, localIP.String(), 7846, "172.21.1.2", 12345, false},
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPanic {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println(test.name, " expected panic occurred ", r)
					}
				}()
				pm := peerManager{conf: test.inCfg}

				pm.init()
			}
		})
	}
}

func Test_peerManager_Stop(t *testing.T) {
	// check if Stop is working.
	tests := []struct {
		name string

		prevStatus int32

		wantStatus      int32
		wantSentChannel bool
	}{
		// never send to finish channel twice.
		{"TInitial", initial, stopping, false},
		{"TRunning", running, stopping, true},
		{"TStopping", stopping, stopping, false},
		{"TStopped", stopped, stopped, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &peerManager{
				logger:        logger,
				finishChannel: make(chan struct{},1 ),
			}

			atomic.StoreInt32(&pm.status, tt.prevStatus)
			pm.Stop()

			if atomic.LoadInt32(&pm.status) != tt.wantStatus {
				t.Errorf("mansger status %v, want %v ", toMStatusName(atomic.LoadInt32(&pm.status)),
					toMStatusName(tt.wantStatus))
			}
			var sent bool
			timeout := time.NewTimer(time.Millisecond<<6)
			select {
				case <-pm.finishChannel:
					sent = true
				case <-timeout.C:
					sent = false
			}
			if sent != tt.wantSentChannel {
				t.Errorf("signal sent %v, want %v ", sent, tt.wantSentChannel)
			}
		})
	}
}

func Test_peerManager_StopInRun(t *testing.T) {
	ctrl := gomock.NewController(t)

	// check if Stop is working.
	tests := []struct {
		name string

		callCnt    int
		wantStatus int32
	}{
		{"TStopOnce", 1, stopped},
		{"TStopTwice", 2, stopped},
		{"TInStopping", 3, stopped},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			mockNT.EXPECT().AddStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()
			mockNT.EXPECT().RemoveStreamHandler(gomock.Any()).AnyTimes()
			pm := &peerManager{
				logger:        logger,
				nt:            mockNT,
				mutex:         &sync.Mutex{},
				awaitDone:     make(chan struct{}),
				finishChannel: make(chan struct{}),
			}
			go pm.runManagePeers()
			// wait status of pm is changed to running
			for atomic.LoadInt32(&pm.status) != running {
				time.Sleep(time.Millisecond)
			}
			// stopping will be done within one second if normal status
			checkTimer := time.NewTimer(time.Second >> 3)
			for i := 0; i < tt.callCnt; i++ {
				pm.Stop()
				time.Sleep(time.Millisecond << 6)
			}
			succ := false
			failedTimeout := time.NewTimer(time.Second * 5)

			// check if status changed
		VERIFYLOOP:
			for {
				select {
				case <-checkTimer.C:
					if atomic.LoadInt32(&pm.status) == tt.wantStatus {
						succ = true
						break VERIFYLOOP
					} else {
						checkTimer.Stop()
						checkTimer.Reset(time.Second)
					}
				case <-failedTimeout.C:
					break VERIFYLOOP
				}
			}
			if !succ {
				t.Errorf("mansger status %v, want %v within %v", toMStatusName(atomic.LoadInt32(&pm.status)),
					toMStatusName(tt.wantStatus), time.Second*5)
			}
		})
	}
}

func toMStatusName(status int32) string {
	switch status {
	case initial:
		return "initial"
	case running:
		return "running"
	case stopping:
		return "stopping"
	case stopped:
		return "stopped"
	default:
		return "(invalid)" + strconv.Itoa(int(status))
	}
}
