/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
)

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	hub *component.ComponentHub

	pm PeerManager
	rm ReconnectManager
}

//var _ component.IComponent = (*P2PComponent)(nil)
var _ ActorService = (*P2P)(nil)

const defaultTTL = time.Second * 4

// NewP2P create a new ActorService for p2p
func NewP2P(hub *component.ComponentHub, cfg *config.Config, chainsvc *blockchain.ChainService) *P2P {

	netsvc := &P2P{
		hub: hub,
	}
	netsvc.BaseComponent = component.NewBaseComponent(message.P2PSvc, netsvc, log.NewLogger("p2p"))
	netsvc.init(cfg, chainsvc)
	return netsvc
}

// Start starts p2p service
func (ns *P2P) BeforeStart() {}

func (ns *P2P) AfterStart() {
	p2ps.pm.Start()
}

// Stop stops
func (p2ps *P2P) BeforeStop() {
	p2ps.pm.Stop()
}

func (p2ps *P2P) Statics() *map[string]interface{} {
	return nil
}

func (p2ps *P2P) init(cfg *config.Config, chainsvc *blockchain.ChainService) {
	reconMan := NewReconnectManager(p2ps.Logger)
	peerMan := NewPeerManager(p2ps, cfg, reconMan, p2ps.Logger)

	// connect managers each other
	reconMan.pm = peerMan

	p2ps.pm = peerMan
	p2ps.rm = reconMan
}

const success bool = true
const failed bool = false

// Receive got actor message and then handle it.
func (p2ps *P2P) Receive(context actor.Context) {

	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case *message.GetAddressesMsg:
		p2ps.GetAddresses(msg.ToWhom, msg.Size)
	case *message.GetBlockHeaders:
		p2ps.GetBlockHeaders(msg)
	case *message.GetBlockInfos:
		p2ps.GetBlocks(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewBlock:
		p2ps.NotifyNewBlock(*msg)
	case *message.GetMissingBlocks:
		p2ps.GetMissingBlocks(msg.ToWhom, msg.Hashes)
	case *message.GetTransactions:
		p2ps.GetTXs(msg.ToWhom, msg.Hashes)
	case *message.NotifyNewTransactions:
		p2ps.NotifyNewTX(*msg)
	case *message.GetPeers:
		peers, states := p2ps.pm.GetPeerAddresses()
		context.Respond(&message.GetPeersRsp{Peers: peers, States: states})
	}
}

// SendRequest implement interface method of ActorService
func (p2ps *P2P) SendRequest(actor string, msg interface{}) {
	p2ps.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (p2ps *P2P) FutureRequest(actor string, msg interface{}) *actor.Future {
	return p2ps.RequestToFuture(actor, msg, defaultTTL)
}

// CallRequest implement interface method of ActorService
func (p2ps *P2P) CallRequest(actor string, msg interface{}) (interface{}, error) {
	future := p2ps.RequestToFuture(actor, msg, defaultTTL)

	return future.Result()
}
