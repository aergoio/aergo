/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"sync"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

func FailTestGetPeers(t *testing.T) {
	mockActorServ := &MockActorService{}
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	mockMF := new(MockMoFactory)
	target := NewPeerManager(nil, nil, mockActorServ,
		cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config),
		nil, nil, new(MockReconnectManager), nil,
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
	mockActorServ := &MockActorService{}
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	mockMF := new(MockMoFactory)

	tLogger := log.NewLogger("test.p2p")
	tConfig := cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config)
	InitNodeInfo(&tConfig.BaseConfig, tConfig.P2P, tLogger)
	target := NewPeerManager(nil, nil, mockActorServ,
		tConfig,
		nil, nil, new(MockReconnectManager), nil,
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
		assert.True(t, cnt > (iterSize>>2))
		waitChan <- 0
	}()

	<-waitChan

	wgAll.Wait()
	assert.True(t, iterSize == len(target.GetPeers()))
}

func TestPeerManager_GetPeerAddresses(t *testing.T) {
	peersLen := 3
	samplePeers := make([]*remotePeerImpl, peersLen)
	samplePeers[0] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID}, lastNotice:&LastBlockStatus{}}
	samplePeers[1] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID2}, lastNotice:&LastBlockStatus{}}
	samplePeers[2] = &remotePeerImpl{meta: p2pcommon.PeerMeta{ID: dummyPeerID3}, lastNotice:&LastBlockStatus{}}
	tests := []struct {
		name string
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pm := &peerManager{remotePeers:make(map[peer.ID]*remotePeerImpl)}
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
	localIP, _ := externalIP()

	tests := []struct {
		name string
		inCfg *cfg.P2PConfig
		expectProtoAddr string
		expectProtoPort uint32
		expectBindAddr string
		expectBindPort uint32
		expectPanic bool
	}{
		{"TDefault",defaultCfg, localIP.String(), uint32(defaultCfg.NetProtocolPort), localIP.String(), uint32(defaultCfg.NetProtocolPort), false},
		// wrong ProtocolAddress 0.0.0.0
		{"TUnspecifiedAddr",&cfg.P2PConfig{NetProtocolAddr:"0.0.0.0",NetProtocolPort:7846}, localIP.String(), 7846, localIP.String(), uint32(defaultCfg.NetProtocolPort), true},
		// wrong ProtocolAddress
		{"TWrongAddr",&cfg.P2PConfig{NetProtocolAddr:"24558.30.0.0",NetProtocolPort:7846}, localIP.String(), 7846, localIP.String(), 7846, true},
		// bind all address
		{"TBindAll",&cfg.P2PConfig{NetProtocolAddr:"",NetProtocolPort:7846, NPBindAddr:"0.0.0.0"}, localIP.String(), 7846, "0.0.0.0", 7846, false},
		// bind differnt address
		{"TBindDifferAddr",&cfg.P2PConfig{NetProtocolAddr:"",NetProtocolPort:7846, NPBindAddr:"172.21.1.2"}, localIP.String(), 7846, "172.21.1.2", 7846, false},
		// bind different port
		{"TDifferPort",&cfg.P2PConfig{NetProtocolAddr:"",NetProtocolPort:7846, NPBindPort:12345}, localIP.String(), 7846, localIP.String(), 12345, false},
		// bind different address and port
		{"TBindDiffer",&cfg.P2PConfig{NetProtocolAddr:"",NetProtocolPort:7846, NPBindAddr:"172.21.1.2", NPBindPort:12345}, localIP.String(), 7846, "172.21.1.2", 12345, false},
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
				pm := peerManager{conf:test.inCfg}

				pm.init()
			}
		})
	}
}

