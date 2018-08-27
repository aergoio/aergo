/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
)

// PingProtocol type
type PingProtocol struct {
	BaseMsgHandler
	actorServ ActorService
	ps        PeerManager

	log      *log.Logger
	reqMutex sync.Mutex
}

// PingHandler handle pingRequest message
type PingHandler struct {
	BaseMsgHandler
}

// NewPingHandler create handler about ping protocol for a peer
func NewPingHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *PingHandler {
	h := &PingHandler{BaseMsgHandler: BaseMsgHandler{protocol: pingRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return h
}

func (p *PingProtocol) setPeerManager(pm PeerManager) {
	p.ps = pm
}

// remote peer requests handler
func (p *PingHandler) handlePing(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer

	pingMsg := &types.Ping{}
	err := unmarshalMessage(msg.Data, pingMsg)
	if err != nil {
		p.logger.Warn().Err(err).Msg("Failed to decode ping message")
		p.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, nil)

	// generate response message
	p.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, msg.Header.Id).Msg("Sending ping response")
	resp := &types.Pong{MessageData: &types.MessageData{}} // BestBlockHash: bestBlock.GetHash(),
	// BestHeight:    bestBlock.GetHeader().GetBlockNo(),

	remotePeer.sendMessage(newPbMsgResponseOrder(msg.Header.Id, false, pingResponse, resp))
}

// remote ping response handler
func (p *PingHandler) handlePingResponse(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	remotePeer := p.peer
	pingRspMsg := &types.Pong{}
	err := unmarshalMessage(msg.Data, pingRspMsg)
	if err != nil {
		p.logger.Warn().Err(err).Msg("Failed to decode ping response message")
		p.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, nil)
	remotePeer.consumeRequest(msg.Header.Id)
}

// remote ping response handler
func (p *PingHandler) handleGoAway(msg *types.P2PMessage) {
	peerID := p.peer.ID()
	goawayMsg := &types.GoAwayNotice{}
	err := unmarshalMessage(msg.Data, goawayMsg)
	if err != nil {
		p.logger.Warn().Err(err).Msg("Failed to decode ping response message")
		p.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(p.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, goawayMsg.Message)
	// TODO: check to remove peer here or not. (the sending peer will disconnect.)
}
