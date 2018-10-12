/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multicodec/protobuf"
)

const aergoP2PSub protocol.ID = "/aergop2p/0.2"

// PeerHandshaker works to handshake to just connected peer
type PeerHandshaker struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID

	localStatus  *types.Status
	remoteStatus *types.Status
}

type hsResult struct {
	statusMsg *types.Status
	err       error
}

func newHandshaker(pm PeerManager, actorServ ActorService, log *log.Logger, peerID peer.ID) *PeerHandshaker {
	return &PeerHandshaker{pm: pm, actorServ: actorServ, logger: log, peerID: peerID}
}

func (h *PeerHandshaker) handshakeOutboundPeerTimeout(rw MsgReadWriter, ttl time.Duration) (*types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		statusMsg, err := h.handshakeOutboundPeer(rw)
		doneChan <- &hsResult{statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, err
	}
	return ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

func (h *PeerHandshaker) handshakeInboundPeerTimeout(rw MsgReadWriter, ttl time.Duration) (*types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		statusMsg, err := h.handshakeInboundPeer(rw)
		doneChan <- &hsResult{statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, err
	}
	return ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

type targetFunc func(chan<- interface{})
type timeoutErr error

func runFuncTimeout(m targetFunc, ttl time.Duration) (interface{}, error) {
	done := make(chan interface{})
	go m(done)
	select {
	case hsResult := <-done:
		return hsResult, nil
	case <-time.NewTimer(ttl).C:
		return nil, fmt.Errorf("timeout").(timeoutErr)
	}
}

// handshakeOutboundPeer start handshake with outbound peer
func (h *PeerHandshaker) handshakeOutboundPeer(rw MsgReadWriter) (*types.Status, error) {

	peerID := h.peerID

	h.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Starting Handshake")
	// send status
	statusMsg, err := createStatusMsg(h.pm, h.actorServ)
	if err != nil {
		return nil, err
	}
	h.localStatus = statusMsg
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

	if err := h.checkProtocolVersion("0.2"); err != nil {
		// h.logger.Info().Err(err).Str(LogPeerID, peerID.Pretty()).Msg("invalid protocol version of peer")
		return nil, err
	}

	if data.Subprotocol() != StatusRequest {
		// TODO: parse message and return
		// h.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", StatusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake response")
		return nil, fmt.Errorf("Unexpected message type")
	}
	statusResp := &types.Status{}
	err = unmarshalMessage(data.Payload(), statusResp)
	if err != nil {
		// h.logger.Warn().Err(err).Msg("Failed to decode status message")
		return nil, err
	}

	h.remoteStatus = statusResp
	// check status message
	return statusResp, nil
}

// onHandshake is handle handshake from inbound peer
func (h *PeerHandshaker) handshakeInboundPeer(rw MsgReadWriter) (*types.Status, error) {
	peerID := h.peerID

	// first message must be status
	data, err := rw.ReadMsg()
	if err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to create p2p message")
		return nil, err
	}

	if err := h.checkProtocolVersion("0.2"); err != nil {
		h.logger.Info().Err(err).Str(LogPeerID, peerID.Pretty()).Msg("invalid protocol version of peer")
		return nil, err
	}

	if data.Subprotocol() != StatusRequest {
		// TODO: parse message and return
		h.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", StatusRequest.String()).Str("actual", data.Subprotocol().String()).Msg("Unexpected message type")
		return nil, fmt.Errorf("Unexpected message type")
	}

	statusMsg := &types.Status{}
	if err := unmarshalMessage(data.Payload(), statusMsg); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("Failed to decode status message")
		return nil, err
	}
	h.remoteStatus = statusMsg

	// send my status message as response
	statusResp, err := createStatusMsg(h.pm, h.actorServ)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to create status message")
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
	h.localStatus = statusResp
	return statusMsg, nil

}

// doPostHandshake is additional work after peer is added.
func (h *PeerHandshaker) doInitialSync() {

	// sync block infos
	h.actorServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: h.peerID, BlockNo: h.remoteStatus.BestHeight, BlockHash: h.remoteStatus.BestBlockHash})

	// sync mempool tx infos
	// TODO add tx handling
}

func createStatusMsg(pm PeerManager, actorServ ActorService) (*types.Status, error) {
	// find my best block
	bestBlock, err := extractBlockFromRequest(actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		return nil, err
	}
	selfAddr := pm.SelfMeta().ToPeerAddress()
	// create message data
	statusMsg := &types.Status{
		Sender:        &selfAddr,
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	return statusMsg, nil
}

func (h *PeerHandshaker) checkProtocolVersion(versionStr string) error {
	// TODO modify interface and put check code here
	return nil
}

func readP2PMessage(rd io.Reader) (*types.P2PMessage, error) {
	containerMsg := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(rd)
	if err := decoder.Decode(containerMsg); err != nil {
		return nil, err
	}
	return containerMsg, nil
}
