/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"io"
)

// V030Handshaker exchange status data over protocol version .0.3.0
type V030Handshaker struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID
	chainID	  *types.ChainID

	rd *bufio.Reader
	wr *bufio.Writer
	msgRW MsgReadWriter
}

type V030HSMessage struct {
	HSHeader
	Sigature [SigLength]byte
	PubKeyB []byte
	Timestamp uint64
	Nonce uint16

}

func (h *V030Handshaker) GetMsgRW() MsgReadWriter {
	return h.msgRW
}


func newV030StateHS(pm PeerManager, actorServ ActorService, log *log.Logger, chainID *types.ChainID, peerID peer.ID, rd io.Reader, wr io.Writer) *V030Handshaker {
	h := &V030Handshaker{pm: pm, actorServ: actorServ, logger: log, chainID:chainID, peerID: peerID, rd: bufio.NewReader(rd), wr:bufio.NewWriter(wr)}
	h.msgRW = NewV030ReadWriter(h.rd, h.wr)
	return h
}

// handshakeOutboundPeer start handshake with outbound peer
func (h *V030Handshaker) doForOutbound() (*types.Status, error) {
	rw := h.msgRW
	peerID := h.peerID

	// TODO need to check auth at first...

	h.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Starting Handshake")
	// send status
	statusMsg, err := createStatusMsg(h.pm, h.actorServ, h.chainID)
	if err != nil {
		return nil, err
	}
	moFactory := &v030MOFactory{}
	container := moFactory.newHandshakeMessage(StatusRequest, statusMsg)
	if container == nil {
		// h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to create p2p message")
		return nil, fmt.Errorf("failed to craete container message")
	}
	if err = rw.WriteMsg(container); err != nil {
		return nil, err
	}

	// and wait to response status
	data, err := rw.ReadMsg()
	if err != nil {
		// h.logger.Info().Err(err).Msg("fail to decode")
		return nil, err
	}

	if data.Subprotocol() != StatusRequest {
		if data.Subprotocol() == GoAway {
			return h.handleGoAway(peerID, data)
		} else {
			return nil, fmt.Errorf("unexpected message type")
		}
	}
	statusResp := &types.Status{}
	err = UnmarshalMessage(data.Payload(), statusResp)
	if err != nil {
		return nil, err
	}

	// check if chainID is same or not
	remoteChainID := types.NewChainID()
	err = remoteChainID.Read(statusResp.ChainID)
	if err != nil {
		return nil, err
	}
	if !h.chainID.Equals(remoteChainID) {
		return nil, fmt.Errorf("different chainID : %s", remoteChainID.ToJSON())
	}

	// check status message
	return statusResp, nil
}

// onConnect is handle handshake from inbound peer
func (h *V030Handshaker) doForInbound() (*types.Status, error) {
	rw := h.msgRW
	peerID := h.peerID

	// TODO need to check auth at first...

	// first message must be status
	data, err := rw.ReadMsg()
	if err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to create p2p message")
		return nil, err
	}

	if data.Subprotocol() != StatusRequest {
		if data.Subprotocol() == GoAway {
			return h.handleGoAway(peerID, data)
		} else {
			h.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", StatusRequest.String()).Str("actual", data.Subprotocol().String()).Msg("unexpected message type")
			return nil, fmt.Errorf("unexpected message type")
		}
	}

	statusMsg := &types.Status{}
	if err := UnmarshalMessage(data.Payload(), statusMsg); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("Failed to decode status message.")
		return nil, err
	}

	// check if chainID is same or not
	remoteChainID := types.NewChainID()
	err = remoteChainID.Read(statusMsg.ChainID)
	if err != nil {
		return nil, err
	}
	if !h.chainID.Equals(remoteChainID) {
		return nil, fmt.Errorf("different chainID : %s", remoteChainID.ToJSON())
	}


	// send my status message as response
	statusResp, err := createStatusMsg(h.pm, h.actorServ, h.chainID)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create status message.")
		return nil, err
	}
	moFactory := &v030MOFactory{}
	container := moFactory.newHandshakeMessage(StatusRequest, statusResp)
	if container == nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Msg("failed to create p2p message")
		return nil, fmt.Errorf("failed to create p2p message")
	}
	if err = rw.WriteMsg(container); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send response status ")
		return nil, err
	}
	return statusMsg, nil

}

func (h *V030Handshaker) handleGoAway(peerID peer.ID, data Message) (*types.Status, error) {
	goAway := &types.GoAwayNotice{}
	if err := UnmarshalMessage(data.Payload(), goAway); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("Remore peer sent goAway but failed to decode internal message")
		return nil, err
	}
	return nil, fmt.Errorf("remote peer refuse handshake: %s",goAway.GetMessage())
}
