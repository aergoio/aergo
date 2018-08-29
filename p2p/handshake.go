/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
)

const aergoP2PSub protocol.ID = "/aergop2p/0.2"

func doHandshake(pm *peerManager, peerID peer.ID, rw *bufio.ReadWriter) bool {
	pm.log.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Starting Handshake")
	// TODO move to caller's function
	if _, found := pm.GetPeer(peerID); found {
		pm.log.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Peer was already added")
		return false
	}

	// send status
	statusMsg, err := createStatusMsg(pm, pm.iServ)
	if err != nil {
		pm.log.Warn().Err(err).Msg("failed to create status message")
		return false
	}
	serialized, err := marshalMessage(statusMsg)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to marshal")
		return false
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.GetMessageData().Subprotocol = statusRequest.Uint32()
	err = SendProtoMessage(container, rw)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send status ")
		return false
	}

	// and wait to response status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(rw)
	err = decoder.Decode(data)
	if err != nil {
		pm.log.Info().Err(err).Msg("fail to decode")
		return false
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		pm.log.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake response")
		return false
	}
	statusResp := &types.Status{}
	err = unmarshalMessage(data.Data, statusResp)
	if err != nil {
		pm.log.Warn().Err(err).Msg("Failed to decode status message")
		return false
	}

	// check status message
	return true
}

func (pm *peerManager) onHandshake(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	rw := &bufio.ReadWriter{Reader: bufio.NewReader(s), Writer: bufio.NewWriter(s)}

	// first message must be status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(s)
	err := decoder.Decode(data)
	if err != nil {
		pm.log.Info().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("fail to decode")
		pm.sendGoAway(rw, "invalid message")
		s.Close()
		return
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		pm.log.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake protocol")
		pm.sendGoAway(rw, "unexpected message type")
		s.Close()
		return
	}
	statusMsg := &types.Status{}
	err = unmarshalMessage(data.Data, statusMsg)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("Failed to decode status message")
		pm.sendGoAway(rw, "invalid status message")
		s.Close()
		return
	}

	// TODO: check status
	meta := FromPeerAddress(statusMsg.Sender)

	// send my status message as response
	statusResp, err := createStatusMsg(pm, pm.iServ)
	if err != nil {
		pm.log.Warn().Err(err).Msg("failed to create status message")
		pm.sendGoAway(rw, "internal error")
		s.Close()
		return
	}
	serialized, err := marshalMessage(statusResp)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to marshal")
		pm.sendGoAway(rw, "internal error")
		s.Close()
		return
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.GetMessageData().Subprotocol = statusRequest.Uint32()

	err = SendProtoMessage(container, rw)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send response status ")
		return
	}

	// try Add peer
	if !pm.tryAddInboundPeer(meta, rw) {
		// failed to add
		pm.sendGoAway(rw, "Concurrent handshake")
		s.Close()
	}
}

func (pm *peerManager) sendGoAway(rw *bufio.ReadWriter, msg string) {
	serialized, err := marshalMessage(&types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg})
	if err != nil {
		pm.log.Warn().Err(err).Msg("failed to marshal")
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.Header.Subprotocol = goAway.Uint32()
	SendProtoMessage(container, rw)
	rw.Flush()
}

func createStatusMsg(ps PeerManager, actorServ ActorService) (*types.Status, error) {
	// find my best block
	bestBlock, err := extractBlockFromRequest(actorServ.CallRequest(message.ChainSvc, &message.GetBestBlock{}))
	if err != nil {
		return nil, err
	}
	selfAddr := ps.SelfMeta().ToPeerAddress()
	// create message data
	statusMsg := &types.Status{
		MessageData:   &types.MessageData{},
		Sender:        &selfAddr,
		BestBlockHash: bestBlock.GetHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	return statusMsg, nil
}
