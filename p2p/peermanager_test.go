/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package p2p

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aergoio/aergo-lib/log"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

// Ignoring test for now, for lack of abstraction on AergoPeer struct
func IgrenoreTestP2PServiceRunAddPeer(t *testing.T) {
	mockActorServ := MockActorService{}
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	target := NewPeerManager(&mockActorServ,
		cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config),
		nil, new(MockReconnectManager),
		log.NewLogger("test.p2p")).(*peerManager)

	target.Host = &mockHost{pstore.NewPeerstore()}
	target.selfMeta.ID = peer.ID("gwegw")
	go target.runManagePeers()

	sampleAddr1 := PeerMeta{ID: "ddd", IPAddress: "192.168.0.1", Port: 33888, Outbound: true}
	sampleAddr2 := PeerMeta{ID: "fff", IPAddress: "192.168.0.2", Port: 33888, Outbound: true}
	target.AddNewPeer(sampleAddr1)
	target.AddNewPeer(sampleAddr1)
	time.Sleep(time.Second)
	if len(target.Peerstore().Peers()) != 1 {
		t.Errorf("Peer count : Expected %d, Actually %d", 1, len(target.Peerstore().Peers()))
	}
	target.AddNewPeer(sampleAddr2)
	time.Sleep(time.Second * 1)
	if len(target.Peerstore().Peers()) != 2 {
		t.Errorf("Peer count : Expected %d, Actually %d", 2, len(target.Peerstore().Peers()))
	}
}

func FailTestGetPeers(t *testing.T) {
	mockActorServ := &MockActorService{}
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)
	target := NewPeerManager(mockActorServ,
		cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config),
		nil, new(MockReconnectManager),
		log.NewLogger("test.p2p")).(*peerManager)

	iterSize := 500
	wg := sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			peerID := peer.ID(strconv.Itoa(i))
			peerMeta := PeerMeta{ID: peerID}
			target.remotePeers[peerID] = newRemotePeer(peerMeta, target, mockActorServ, logger, nil, nil)
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

func TestGetPeers(t *testing.T) {
	mockActorServ := &MockActorService{}
	dummyBlock := types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActorServ.On("CallRequest", mock.Anything, mock.Anything).Return(message.GetBlockRsp{Block: &dummyBlock}, nil)

	tLogger := log.NewLogger("test.p2p")
	tConfig := cfg.NewServerContext("", "").GetDefaultConfig().(*cfg.Config)
	InitNodeInfo(tConfig.P2P, tLogger)
	target := NewPeerManager(mockActorServ,
		tConfig,
		nil, new(MockReconnectManager),
		tLogger).(*peerManager)

	iterSize := 500
	wg := &sync.WaitGroup{}
	wgAll := &sync.WaitGroup{}
	waitChan := make(chan int)
	wg.Add(1)
	wgAll.Add(1)
	go func() {
		for i := 0; i < iterSize; i++ {
			peerID := peer.ID(strconv.Itoa(i))
			peerMeta := PeerMeta{ID: peerID}
			target.insertPeer(peerID, newRemotePeer(peerMeta, target, mockActorServ, logger, nil, nil))
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
