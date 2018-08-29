/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"strconv"
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

// initiateHandshake start handshake with outbound peer
func initiateHandshake(pm *peerManager, peerID peer.ID, rw *bufio.ReadWriter) (*types.Status, bool) {
	pm.log.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Starting Handshake")
	// TODO move to caller's function
	if _, found := pm.GetPeer(peerID); found {
		pm.log.Debug().Str(LogPeerID, peerID.Pretty()).Msg("Peer was already added")
		return nil, false
	}

	// send status
	statusMsg, err := createStatusMsg(pm, pm.iServ)
	if err != nil {
		pm.log.Warn().Err(err).Msg("failed to create status message")
		return nil, false
	}
	serialized, err := marshalMessage(statusMsg)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to marshal")
		return nil, false
	}
	container := &types.P2PMessage{Header: &types.MessageData{}, Data: serialized}
	setupMessageData(container.Header, uuid.Must(uuid.NewV4()).String(), false, ClientVersion, time.Now().Unix())
	container.GetMessageData().Subprotocol = statusRequest.Uint32()
	err = SendProtoMessage(container, rw)
	if err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send status ")
		return nil, false
	}

	// and wait to response status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(rw)
	err = decoder.Decode(data)
	if err != nil {
		pm.log.Info().Err(err).Msg("fail to decode")
		return nil, false
	}

	if data.Header.GetSubprotocol() != statusRequest.Uint32() {
		// TODO: parse message and return
		pm.log.Info().Str(LogPeerID, peerID.Pretty()).Str("expected", statusRequest.String()).Str("actual", SubProtocol(data.Header.GetSubprotocol()).String()).Msg("Unexpected handshake response")
		return nil, false
	}
	statusResp := &types.Status{}
	err = unmarshalMessage(data.Data, statusResp)
	if err != nil {
		pm.log.Warn().Err(err).Msg("Failed to decode status message")
		return nil, false
	}

	// check status message
	return statusResp, true
}

// onHandshake is handle handshake from inbound peer
func (pm *peerManager) onHandshake(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	rw := &bufio.ReadWriter{Reader: bufio.NewReader(s), Writer: bufio.NewWriter(s)}

	// first message must be status
	data := &types.P2PMessage{}
	decoder := mc_pb.Multicodec(nil).Decoder(s)
	if err := decoder.Decode(data); err != nil {
		pm.log.Info().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("fail to decode")
		pm.sendGoAway(rw, "invalid message")
		s.Close()
		return
	}

	if err := pm.checkProtocolVersion(); err != nil {
		pm.log.Info().Err(err).Str(LogPeerID, peerID.Pretty()).Msg("invalid protocol version of peer")
		pm.sendGoAway(rw, "Handshake failed")
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
	if err := unmarshalMessage(data.Data, statusMsg); err != nil {
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

	if err = SendProtoMessage(container, rw); err != nil {
		pm.log.Warn().Str(LogPeerID, peerID.Pretty()).Err(err).Msg("failed to send response status ")
		return
	}

	// try Add peer
	if !pm.tryAddInboundPeer(meta, rw) {
		// failed to add
		pm.sendGoAway(rw, "Concurrent handshake")
		s.Close()
	}

	pm.iServ.SendRequest(message.ChainSvc, &message.SyncBlockState{PeerID: peerID, BlockNo: statusMsg.BestHeight, BlockHash: statusMsg.BestBlockHash})

	// notice to p2pmanager that handshaking is finished
	pm.NotifyPeerHandshake(peerID)
}

func (ps *peerManager) tryAddInboundPeer(meta PeerMeta, rw *bufio.ReadWriter) bool {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	peerID := meta.ID
	peer, found := ps.remotePeers[peerID]

	if found {
		// already found. drop this connection
		if ComparePeerID(ps.selfMeta.ID, peerID) <= 0 {
			return false
		}
	}
	peer = newRemotePeer(meta, ps, ps.iServ, ps.log)
	peer.rw = rw
	ps.insertHandlers(peer)
	go peer.runPeer()
	peer.setState(types.RUNNING)
	ps.insertPeer(peerID, peer)
	peerAddr := meta.ToPeerAddress()
	ps.log.Info().Str(LogPeerID, peerID.Pretty()).Str("addr", getIP(&peerAddr).String()+":"+strconv.Itoa(int(peerAddr.Port))).Msg("Inbound peer is  added to peerService")
	return true
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
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
	}

	return statusMsg, nil
}

func (pm *peerManager) checkProtocolVersion() error {
	// TODO modify interface and put check code here
	return nil
}
