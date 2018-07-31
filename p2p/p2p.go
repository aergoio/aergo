/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"reflect"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
)

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	hub *component.ComponentHub

	p2ps PeerManager

	ping  *PingProtocol
	addrs *AddressesProtocol
	blk   *BlockProtocol
	txs   *TxProtocol
}

//var _ component.IComponent = (*P2PComponent)(nil)
var _ ActorService = (*P2P)(nil)

const defaultTTL = time.Second * 4

// NewP2P create a new ActorService for p2p
func NewP2P(hub *component.ComponentHub, cfg *config.Config, chainsvc *blockchain.ChainService) *P2P {

	netsvc := &P2P{
		BaseComponent: component.NewBaseComponent(message.P2PSvc, log.NewLogger(log.P2PSvc), cfg.EnableDebugMsg),
		hub:           hub,
	}

	netsvc.init(cfg, chainsvc)
	return netsvc
}

// Start starts p2p service
func (ns *P2P) Start() {
	ns.BaseComponent.Start(ns)
	//TODO add init logics for this service
	ns.p2ps.Start()
}

// Stop stops
func (ns *P2P) Stop() {
	ns.p2ps.Stop()

	ns.BaseComponent.Stop()
}

func (ns *P2P) init(cfg *config.Config, chainsvc *blockchain.ChainService) PeerManager {
	p2psvc := NewPeerManager(ns, cfg, ns.Logger)
	// FIXME 초기화
	ns.ping = NewPingProtocol(ns.Logger)
	ns.ping.actorServ = ns
	p2psvc.AddSubProtocol(ns.ping)

	ns.blk = NewBlockProtocol(ns.Logger, chainsvc)
	ns.blk.iserv = ns
	ns.blk.log = ns.Logger
	p2psvc.AddSubProtocol(ns.blk)

	ns.addrs = NewAddressesProtocol(ns.Logger)
	p2psvc.AddSubProtocol(ns.addrs)

	ns.txs = NewTxProtocol(ns.Logger, chainsvc)
	ns.txs.iserv = ns
	p2psvc.AddSubProtocol(ns.txs)

	ns.p2ps = p2psvc
	return p2psvc
}

const success bool = true
const failed bool = false

// Receive got actor message and then handle it.
func (ns *P2P) Receive(context actor.Context) {
	ns.BaseComponent.Receive(context)

	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	// case *message.PingMsg:
	// 	result := ns.ping.Ping(msg.ToWhom)
	// 	context.Respond(result)
	case *message.GetAddressesMsg:
		ns.addrs.GetAddresses(msg.ToWhom, msg.Size)
	case *message.GetBlockHeaders:
		ns.blk.GetBlockHeaders(msg)
	case *message.GetBlockInfos:
		ns.blk.GetBlocks(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewBlock:
		ns.blk.NotifyNewBlock(*msg)
	case *message.GetTransactions:
		ns.txs.GetTXs(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewTransactions:
		ns.txs.NotifyNewTX(*msg)
	case *message.GetPeers:
		peers := ns.p2ps.GetPeerAddresses()
		context.Respond(&message.GetPeersRsp{Peers: peers})
	case *message.GetMissingBlocks:
		ns.blk.GetMissingBlocks(msg.ToWhom, msg.Hashes)
	default:
		ns.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}

// SendRequest implement interface method of ActorService
func (ns *P2P) SendRequest(actor string, msg interface{}) {
	ns.hub.Request(actor, msg, ns)
}

// FutureRequest implement interface method of ActorService
func (ns *P2P) FutureRequest(actor string, msg interface{}) *actor.Future {
	return ns.hub.RequestFuture(actor, msg, defaultTTL)
}

// CallRequest implement interface method of ActorService
func (ns *P2P) CallRequest(actor string, msg interface{}) (interface{}, error) {
	future := ns.hub.RequestFuture(actor, msg, defaultTTL)

	return future.Result()
}
