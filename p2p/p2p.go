/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"io/ioutil"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

type nodeInfo struct {
	id      peer.ID
	sid     string
	pubKey  crypto.PubKey
	privKey crypto.PrivKey
}

// P2P is actor component for p2p
type P2P struct {
	*component.BaseComponent

	hub *component.ComponentHub

	pm     PeerManager
	rm     ReconnectManager
	signer msgSigner
}

var (
	_  ActorService = (*P2P)(nil)
	ni *nodeInfo
)

// InitNodeInfo initializes node-specific informations like node id.
// Caution: this must be called before all the goroutines are started.
func InitNodeInfo(cfg *config.P2PConfig, logger *log.Logger) {
	// check Key and address
	var (
		priv crypto.PrivKey
		pub  crypto.PubKey
	)

	if cfg.NPKey != "" {
		dat, err := ioutil.ReadFile(cfg.NPKey)
		if err == nil {
			priv, err = crypto.UnmarshalPrivateKey(dat)
			if err != nil {
				logger.Warn().Str("npkey", cfg.NPKey).Msg("invalid keyfile. It's not private key file")
			}
			pub = priv.GetPublic()
		} else {
			logger.Warn().Str("npkey", cfg.NPKey).Msg("invalid keyfile path")
		}
	}

	if priv == nil {
		logger.Info().Msg("No valid private key file is found. use temporary pk instead")
		priv, pub, _ = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	}

	id, _ := peer.IDFromPublicKey(pub)

	ni = &nodeInfo{
		id:      id,
		sid:     enc.ToString([]byte(id)),
		pubKey:  pub,
		privKey: priv,
	}
}

// NodeID returns the node id.
func NodeID() peer.ID {
	return ni.id
}

// NodeSID returns the string representation of the node id.
func NodeSID() string {
	return ni.sid
}

// NodePrivKey returns the private key of the node.
func NodePrivKey() crypto.PrivKey {
	return ni.privKey
}

// NodePubKey returns the public key of the node.
func NodePubKey() crypto.PubKey {
	return ni.pubKey
}

// NewP2P create a new ActorService for p2p
func NewP2P(hub *component.ComponentHub, cfg *config.Config, chainsvc *blockchain.ChainService) *P2P {
	p2psvc := &P2P{
		hub: hub,
	}
	p2psvc.BaseComponent = component.NewBaseComponent(message.P2PSvc, p2psvc, log.NewLogger("p2p"))
	p2psvc.init(cfg, chainsvc)
	return p2psvc
}

// BeforeStart starts p2p service.
func (p2ps *P2P) BeforeStart() {}

func (p2ps *P2P) AfterStart() {
	if err := p2ps.pm.Start(); err != nil {
		panic("Failed to start p2p component")
	}
}

// BeforeStop is called before actor hub stops. it finishes underlying peer manager
func (p2ps *P2P) BeforeStop() {
	if err := p2ps.pm.Stop(); err != nil {
		p2ps.Logger.Warn().Err(err).Msg("Erro on stopping peerManager")
	}
}

// Statics show statistic information of p2p module. NOTE: It it not implemented yet
func (p2ps *P2P) Statics() *map[string]interface{} {
	return nil
}

func (p2ps *P2P) init(cfg *config.Config, chainsvc *blockchain.ChainService) {
	signer := newDefaultMsgSigner(ni.privKey, ni.pubKey, ni.id)
	reconMan := newReconnectManager(p2ps.Logger)
	peerMan := NewPeerManager(p2ps, cfg, signer, reconMan, p2ps.Logger)
	// connect managers each other
	reconMan.pm = peerMan

	p2ps.signer = signer
	p2ps.pm = peerMan
	p2ps.rm = reconMan
}

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
