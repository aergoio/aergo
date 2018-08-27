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
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multicodec/protobuf"
	uuid "github.com/satori/go.uuid"
)

const aergoP2PSub protocol.ID = "/aergop2p/0.2"

func doHandshake(p *peerManager, rw *bufio.ReadWriter) bool {
	// TODO move to caller's function
	if _, found := p.GetPeer(p.ID()); found {
		p.log.Debug().Str(LogPeerID, p.ID().Pretty()).Msg("Peer was already added")
		return false
	}

	// send status
	statusMsg, err := createStatusMsg(p, p.iServ)
	if err != nil {
		p.log.Warn().Err(err).Msg("failed to create status message")
		return false
	}
	serialized, err := marshalMessage(statusMsg)
	if err != nil {
		p.log.Warn().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("failed to marshal")
		return false
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.GetMessageData().Subprotocol = statusRequest.Uint32()
	err = SendProtoMessage(container, rw)
	if err != nil {
		p.log.Warn().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("failed to send status ")
		return false
	}

	// and wait to response status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(rw)
	err = decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		return false
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		p.log.Info().Str(LogPeerID, p.ID().Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake response")
		return false
	}
	statusResp := &types.Status{}
	err = unmarshalMessage(data.Data, statusResp)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode status message")
		return false
	}

	// check status message
	return true
}

func (p *peerManager) onHandshake(s inet.Stream) {
	rw := &bufio.ReadWriter{Reader: bufio.NewReader(s), Writer: bufio.NewWriter(s)}

	// first message must be status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(s)
	err := decoder.Decode(data)
	if err != nil {
		p.log.Info().Err(err).Msg("fail to decode")
		p.sendGoAway(rw, "invalid message")
		s.Close()
		return
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		p.log.Info().Str(LogPeerID, p.ID().Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake protocol")
		p.sendGoAway(rw, "unexpected message type")
		s.Close()
		return
	}
	statusMsg := &types.Status{}
	err = unmarshalMessage(data.Data, statusMsg)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode status message")
		p.sendGoAway(rw, "invalid status message")
		s.Close()
		return
	}

	// TODO: check status
	meta := FromPeerAddress(statusMsg.Sender)

	// send my status message as response
	statusResp, err := createStatusMsg(p, p.iServ)
	if err != nil {
		p.log.Warn().Err(err).Msg("failed to create status message")
		p.sendGoAway(rw, "internal error")
		s.Close()
		return
	}
	serialized, err := marshalMessage(statusResp)
	if err != nil {
		p.log.Warn().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("failed to marshal")
		p.sendGoAway(rw, "internal error")
		s.Close()
		return
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.GetMessageData().Subprotocol = statusRequest.Uint32()

	err = SendProtoMessage(container, rw)
	if err != nil {
		p.log.Warn().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("failed to send response status ")
		return
	}

	// try Add peer
	if !p.tryAddInboundPeer(meta, rw) {
		// failed to add
		p.sendGoAway(rw, "Concurrent handshake")
		s.Close()
	}
}

func (p *peerManager) sendGoAway(rw *bufio.ReadWriter, msg string) {
	serialized, err := marshalMessage(&types.GoAwayNotice{MessageData: &types.MessageData{}, Message: msg})
	if err != nil {
		p.log.Warn().Str(LogPeerID, p.ID().Pretty()).Err(err).Msg("failed to marshal")
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
