/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/p2p/transport"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
)

// P2P is actor component for p2p
type LiteContainerService struct {
	*component.BaseComponent

	dummySetting p2pcommon.LocalSettings
	chainID      *types.ChainID
	meta         p2pcommon.PeerMeta
	nt           p2pcommon.NetworkTransport

	mutex sync.Mutex
}

func (lntc *LiteContainerService) ConsensusAccessor() consensus.ConsensusAccessor {
	panic("implement me")
}

func (lntc *LiteContainerService) PeerManager() p2pcommon.PeerManager {
	panic("implement me")
}

func (lntc *LiteContainerService) LocalSettings() p2pcommon.LocalSettings {
	return lntc.dummySetting
}

func (lntc *LiteContainerService) RoleManager() p2pcommon.PeerRoleManager {
	panic("implement me")
}

var (
// _ ActorService     = (*LiteContainerService)(nil)
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

func (lntc *LiteContainerService) GetNetworkTransport() p2pcommon.NetworkTransport {
	lntc.mutex.Lock()
	defer lntc.mutex.Unlock()
	return lntc.nt
}

func (lntc *LiteContainerService) GenesisChainID() *types.ChainID {
	return lntc.chainID
}

func (lntc *LiteContainerService) init(cfg *config.Config) {
	// load genesis file
	// init from genesis file

	genesis, err := readGenesis(cfg.Polaris.GenesisFile)
	if err != nil {
		panic(err.Error())
	}
	chainIdBytes, err := genesis.ChainID()
	if err != nil {
		panic("genesis block is not set properly: " + err.Error())
	}
	chainID := types.NewChainID()
	err = chainID.Read(chainIdBytes)
	if err != nil {
		panic("invalid chainid: " + err.Error())
	}
	lntc.chainID = chainID

	lntc.meta = initMeta(p2pkey.NodeID(), cfg.P2P)
	lntc.Logger.Info().Str("genesis", chainID.ToJSON()).Msg("genesis block loaded")

	netTransport := transport.NewNetworkTransport(cfg.P2P, lntc.Logger, lntc)

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
	selfPeerID := p2pkey.NodeID()
	peerMetas := make([]p2pcommon.PeerMeta, 0, len(peers))
	for _, rPeerAddr := range peers {
		rPeerID := types.PeerID(rPeerAddr.PeerID)
		if selfPeerID == rPeerID {
			continue
		}
		meta := p2pcommon.FromPeerAddress(rPeerAddr)
		peerMetas = append(peerMetas, meta)
	}
}

func (lntc *LiteContainerService) SelfMeta() p2pcommon.PeerMeta {
	return lntc.meta
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
	return lntc.RequestToFuture(actor, msg, p2pcommon.DefaultActorMsgTTL)
}

// CallRequest implement interface method of ActorService
func (lntc *LiteContainerService) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := lntc.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (lntc *LiteContainerService) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := lntc.RequestToFuture(actor, msg, p2pcommon.DefaultActorMsgTTL)
	return future.Result()
}

func (lntc *LiteContainerService) SelfNodeID() types.PeerID {
	return lntc.meta.ID
}

func (lntc *LiteContainerService) SelfRole() types.PeerRole {
	// return dummy value
	return types.PeerRole_Watcher
}

func (lntc *LiteContainerService) GetChainAccessor() types.ChainAccessor {
	// return dummy value
	return nil
}

func (lntc *LiteContainerService) CertificateManager() p2pcommon.CertificateManager {
	// return dummy value
	return nil
}

// it is copy of initMeta() in p2p package
func initMeta(peerID types.PeerID, conf *config.P2PConfig) p2pcommon.PeerMeta {
	protocolAddr := conf.NetProtocolAddr
	var ipAddress net.IP
	var err error
	var protocolPort int
	if len(conf.NetProtocolAddr) != 0 {
		ipAddress, err = network.GetSingleIPAddress(protocolAddr)
		if err != nil {
			panic("Invalid protocol address " + protocolAddr + " : " + err.Error())
		}
		if ipAddress.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := p2putil.ExternalIP()
		if err != nil {
			panic("error while finding IP address: " + err.Error())
		}
		ipAddress = extIP
		protocolAddr = ipAddress.String()
	}
	protocolPort = conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(conf.NetProtocolPort))
	}
	ma, err := types.ToMultiAddr(ipAddress.String(), uint32(protocolPort))
	var meta p2pcommon.PeerMeta
	meta.ID = peerID
	meta.Addresses = []types.Multiaddr{ma}
	meta.Hidden = !conf.NPExposeSelf
	meta.Version = p2pkey.NodeVersion()

	return meta
}
