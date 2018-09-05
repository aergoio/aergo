/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
)

const aergoP2PSub protocol.ID = "/aergop2p/0.2"

// PeerHandshaker works to handshake to just connected peer
type PeerHandshaker struct {
	pm        PeerManager
	actorServ ActorService
	logger    *log.Logger
	peerID    peer.ID
}

type hsResult struct {
	statusMsg *types.Status
	err       error
}

func newHandshaker(pm PeerManager, actorServ ActorService, log *log.Logger, peerID peer.ID) *PeerHandshaker {
	return &PeerHandshaker{pm: pm, actorServ: actorServ, logger: log, peerID: peerID}
}

func (h *PeerHandshaker) handshakeOutboundPeerTimeout(rw *bufio.ReadWriter, ttl time.Duration) (*types.Status, error) {
	ret, err := runFuncTimeout(func(doneChan chan<- interface{}) {
		statusMsg, err := h.handshakeOutboundPeer(rw)
		doneChan <- &hsResult{statusMsg: statusMsg, err: err}
	}, ttl)
	if err != nil {
		return nil, err
	}
	return ret.(*hsResult).statusMsg, ret.(*hsResult).err
}

func (h *PeerHandshaker) handshakeInboundPeerTimeout(rw *bufio.ReadWriter, ttl time.Duration) (*types.Status, error) {
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
func (h *PeerHandshaker) handshakeOutboundPeer(rw *bufio.ReadWriter) (*types.Status, error) {
	peerID := h.peerID

	h.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Starting Handshake")
	// send status
	statusMsg, err := createStatusMsg(h.pm, h.actorServ)
	if err != nil {
		return nil, err
	}
	container := newP2PMessage("", false, statusRequest, statusMsg)
	if container == nil {
		// h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to create p2p message")
		return nil, fmt.Errorf("failed to craete container message")
	}
	if err = SendProtoMessage(container, rw); err != nil {
		return nil, err
	}

	// and wait to response status
	data, err := readP2PMessage(rw)
	if err != nil {
		// h.logger.Info().Err(err).Msg("fail to decode")
		return nil, err
	}

	if err := h.checkProtocolVersion("0.2"); err != nil {
		// h.logger.Info().Err(err).Str(LogPeerID, peerID.Pretty()).Msg("invalid protocol version of peer")
		return nil, err
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		// h.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake response")
		return nil, fmt.Errorf("Unexpected message type")
	}
	statusResp := &types.Status{}
	err = unmarshalMessage(data.Data, statusResp)
	if err != nil {
		// h.logger.Warn().Err(err).Msg("Failed to decode status message")
		return nil, err
	}

	// check status message
	return statusResp, nil
}

// onHandshake is handle handshake from inbound peer
func (h *PeerHandshaker) handshakeInboundPeer(rw *bufio.ReadWriter) (*types.Status, error) {
	peerID := h.peerID

	// first message must be status
	data, err := readP2PMessage(rw)
	if err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to create p2p message")
		return nil, err
	}

	if err := h.checkProtocolVersion("0.2"); err != nil {
		h.logger.Info().Err(err).Str(LogPeerID, peerID.Pretty()).Msg("invalid protocol version of peer")
		return nil, err
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		h.logger.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected message type")
		return nil, fmt.Errorf("Unexpected message type")
	}

	statusMsg := &types.Status{}
	if err := unmarshalMessage(data.Data, statusMsg); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("Failed to decode status message")
		return nil, err
	}

	// send my status message as response
	statusResp, err := createStatusMsg(h.pm, h.actorServ)
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to create status message")
		return nil, err
	}
	container := newP2PMessage("", false, statusRequest, statusResp)
	if container == nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Msg("failed to create p2p message")
		return nil, fmt.Errorf("failed to create p2p message")
	}
	if err = SendProtoMessage(container, rw); err != nil {
		h.logger.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send response status ")
		return nil, err
	}
	return statusMsg, nil

}

func (h *PeerHandshaker) sendGoAway(rw *bufio.ReadWriter, msg string) {
	serialized, err := marshalMessage(&types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg})
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to marshal")
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.Header.Subprotocol = goAway.Uint32()
	SendProtoMessage(container, rw)
	rw.Flush()
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
		MessageData:   &types.MessageData{},
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
