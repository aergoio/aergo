/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

// P2P is actor component for p2p
type LiteContainerService struct {
	*component.BaseComponent

	chainID *types.ChainID
	nt     p2p.NetworkTransport

	mutex sync.Mutex
}

var (
	//_ ActorService     = (*LiteContainerService)(nil)
)

// NewP2P create a new ActorService for p2p
func NewNTContainer(cfg *config.Config) *LiteContainerService {
	lntc := &LiteContainerService{}
	lntc.BaseComponent = component.NewBaseComponent(message.P2PSvc, lntc, log.NewLogger("p2p"))
	lntc.init(cfg)
	return lntc
}


func (lntc *LiteContainerService) SetHub(hub *component.ComponentHub) {
	lntc.BaseComponent.SetHub(hub)
}

// BeforeStart starts p2p service.
func (lntc *LiteContainerService) BeforeStart() {}

func (lntc *LiteContainerService) AfterStart() {
	lntc.mutex.Lock()
	nt := lntc.nt
	nt.Start()
	lntc.mutex.Unlock()
}

// BeforeStop is called before actor hub stops. it finishes underlying peer manager
func (lntc *LiteContainerService) BeforeStop() {
	lntc.mutex.Lock()
	nt := lntc.nt
	lntc.mutex.Unlock()
	nt.Stop()
}

// Statistics show statistic information of p2p module. NOTE: It it not implemented yet
func (lntc *LiteContainerService) Statistics() *map[string]interface{} {
	return nil
}

func (lntc *LiteContainerService) GetNetworkTransport() p2p.NetworkTransport {
	lntc.mutex.Lock()
	defer lntc.mutex.Unlock()
	return lntc.nt
}

func (lntc *LiteContainerService) ChainID() *types.ChainID {
	return lntc.chainID
}

func (lntc *LiteContainerService) init(cfg *config.Config) {
	// load genesis file
	// init from genesis file
	// TODO code duplication. refactor to delete dupplicate with p2p.go
	genesis, err := readGenesis(cfg.Polaris.GenesisFile)
	if err != nil {
		panic(err.Error())
	}
	chainIdBytes, err := genesis.ChainID()
	if err != nil {
		panic("genesis block is not set properly: "+err.Error())
	}
	chainID := types.NewChainID()
	err = chainID.Read(chainIdBytes)
	if err != nil {
		panic("invalid chainid: "+err.Error())
	}
	lntc.chainID = chainID

	lntc.Logger.Info().Str("genesis",chainID.ToJSON()).Msg("genesis block loaded")

	netTransport := p2p.NewNetworkTransport(cfg.P2P, lntc.Logger)

	lntc.mutex.Lock()
	lntc.nt = netTransport
	lntc.mutex.Unlock()
}

// Receive got actor message and then handle it.
func (lntc *LiteContainerService) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case time.Time:
		lntc.Logger.Debug().Interface("time", msg.String()).Msg("why time is came?")
	default:
		lntc.Logger.Debug().Interface("type", msg).Msg("unexpected msg was sent")
		// do nothing
	}
}

// TODO need refactoring. this code is copied from subprotcoladdrs.go
func (lntc *LiteContainerService) checkAndAddPeerAddresses(peers []*types.PeerAddress) {
	selfPeerID := lntc.nt.SelfNodeID()
	peerMetas := make([]p2pcommon.PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := peer.ID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		meta := p2pcommon.FromPeerAddress(rPeerAddr)
		peerMetas = append(peerMetas, meta)
	}
}

// TellRequest implement interface method of ActorService
func (lntc *LiteContainerService) TellRequest(actor string, msg interface{}) {
	lntc.TellTo(actor, msg)
}

// SendRequest implement interface method of ActorService
func (lntc *LiteContainerService) SendRequest(actor string, msg interface{}) {
	lntc.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (lntc *LiteContainerService) FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future {
	return lntc.RequestToFuture(actor, msg, timeout)
}

// FutureRequestDefaultTimeout implement interface method of ActorService
func (lntc *LiteContainerService) FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future {
	return lntc.RequestToFuture(actor, msg, p2p.DefaultActorMsgTTL)
}

// CallRequest implement interface method of ActorService
func (lntc *LiteContainerService) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := lntc.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (lntc *LiteContainerService) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := lntc.RequestToFuture(actor, msg, p2p.DefaultActorMsgTTL)
	return future.Result()
}
