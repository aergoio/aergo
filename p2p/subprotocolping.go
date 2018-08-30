/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
)

// PingHandler handle pingRequest message
type PingHandler struct {
	BaseMsgHandler
}

// NewPingHandler create handler about ping protocol for a peer
func NewPingHandler(pm PeerManager, peer *RemotePeer, logger *log.Logger) *PingHandler {
	ph := &PingHandler{BaseMsgHandler: BaseMsgHandler{protocol: pingRequest, pm: pm, peer: peer, actor: peer.actorServ, logger: logger}}
	return ph
}

// remote peer requests handler
func (ph *PingHandler) handlePing(msg *types.P2PMessage) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer

	pingMsg := &types.Ping{}
	err := unmarshalMessage(msg.Data, pingMsg)
	if err != nil {
		ph.logger.Warn().Err(err).Msg("Failed to decode ping message")
		ph.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(ph.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, nil)

	// generate response message
	ph.logger.Debug().Str(LogPeerID, peerID.Pretty()).Str(LogMsgID, msg.Header.Id).Msg("Sending ping response")
	resp := &types.Pong{MessageData: &types.MessageData{}} // BestBlockHash: bestBlock.GetHash(),
	// BestHeight:    bestBlock.GetHeader().GetBlockNo(),

	remotePeer.sendMessage(newPbMsgResponseOrder(msg.Header.Id, false, pingResponse, resp))
}

// remote ping response handler
func (ph *PingHandler) handlePingResponse(msg *types.P2PMessage) {
	peerID := ph.peer.ID()
	remotePeer := ph.peer
	pingRspMsg := &types.Pong{}
	err := unmarshalMessage(msg.Data, pingRspMsg)
	if err != nil {
		ph.logger.Warn().Err(err).Msg("Failed to decode ping response message")
		ph.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(ph.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, nil)
	remotePeer.consumeRequest(msg.Header.Id)
}

// remote ping response handler
func (ph *PingHandler) handleGoAway(msg *types.P2PMessage) {
	peerID := ph.peer.ID()
	goawayMsg := &types.GoAwayNotice{}
	err := unmarshalMessage(msg.Data, goawayMsg)
	if err != nil {
		ph.logger.Warn().Err(err).Msg("Failed to decode ping response message")
		ph.peer.sendGoAway("invalid protocol message")
		return
	}
	debugLogReceiveMsg(ph.logger, SubProtocol(msg.Header.Subprotocol), msg.Header.Id, peerID, goawayMsg.Message)
	// TODO: check to remove peer here or not. (the sending peer will disconnect.)
}
