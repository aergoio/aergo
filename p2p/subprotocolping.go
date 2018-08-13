/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"sync"

	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"

	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/multiformats/go-multicodec/protobuf"
)

// pattern: /protocol-name/request-or-response-message/version
const (
	pingRequest   protocol.ID = "/ping/pingreq/0.1"
	pingResponse  protocol.ID = "/ping/pingresp/0.1"
	statusRequest protocol.ID = "/ping/status/0.1"
	goAway        protocol.ID = "/ping/goaway/0.1"
)

// PingProtocol type
type PingProtocol struct {
	actorServ ActorService
	ps        PeerManager

	log      *log.Logger
	reqMutex sync.Mutex
}

// NewPingProtocol create ping subprotocol
func NewPingProtocol(logger *log.Logger) *PingProtocol {
	p := &PingProtocol{log: logger,
		reqMutex: sync.Mutex{}}
	return p
}

func (p *PingProtocol) initWith(p2pservice PeerManager) {
	p.ps = p2pservice
	p.ps.SetStreamHandler(pingRequest, p.onPingRequest)
	p.ps.SetStreamHandler(pingResponse, p.onPingResponse)
	p.ps.SetStreamHandler(statusRequest, p.onStatusRequest)
	p.ps.SetStreamHandler(goAway, p.onGoaway)
}

// remote peer requests handler
func (p *PingProtocol) onPingRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	// get request data
	data := &types.Ping{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode ping request")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)

	// generate response message
	p.log.Debug().Msgf("Sending pong to %s. MSG ID %s.", s.Conn().RemotePeer().Pretty(), data.MessageData.Id)
	resp := &types.Pong{MessageData: &types.MessageData{}} // BestBlockHash: bestBlock.GetHash(),
	// BestHeight:    bestBlock.GetHeader().GetBlockNo(),

	remotePeer.sendMessage(newPbMsgResponseOrder(data.MessageData.Id, false, pingResponse, resp))
}

// remote ping response handler
func (p *PingProtocol) onPingResponse(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	remotePeer, ok := p.ps.GetPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()
	perr := remotePeer.checkState()
	if perr != nil {
		p.log.Info().Msgf("%s: Invalid peer state to handle request %s : %s", peerID.Pretty(), s.Protocol(), perr.Error())
		return
	}

	data := &types.Pong{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode pong request")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	remotePeer.consumeRequest(data.MessageData.Id)
}

func (p *PingProtocol) onStatusRequest(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	p.log.Debug().Str("peer_id", peerID.Pretty()).Msg("Got status message")
	remotePeer, ok := p.ps.LookupPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.Status{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode ping request")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Warn().Msg("Failed to authenticate message")
		return
	}
	p.log.Debug().Str("peer_id", peerID.Pretty()).Msg("starting handshake")
	remotePeer.op <- OpOrder{op: OpHandleHS, param1: data}
}

func (p *PingProtocol) onGoaway(s inet.Stream) {
	peerID := s.Conn().RemotePeer()
	p.log.Debug().Str("peer_id", peerID.Pretty()).Msg("Got goaway message")
	remotePeer, ok := p.ps.LookupPeer(peerID)
	if !ok {
		warnLogUnknownPeer(p.log, s.Protocol(), peerID)
		return
	}

	remotePeer.readLock.Lock()
	defer remotePeer.readLock.Unlock()

	// get request data
	data := &types.GoAwayNotice{}
	decoder := mc_pb.Multicodec(nil).Decoder(bufio.NewReader(s))
	err := decoder.Decode(data)
	if err != nil {
		p.log.Warn().Err(err).Msg("Failed to decode goaway request")
		return
	}
	debugLogReceiveMsg(p.log, s.Protocol(), data.MessageData.Id, peerID, nil)
	valid := p.ps.AuthenticateMessage(data, data.MessageData)
	if !valid {
		p.log.Warn().Msg("Failed to authenticate message")
		return
	}

	p.log.Debug().Msgf("Remote Peer %s kick out me: %s ", peerID.Pretty(), data.Message)
	p.ps.RemovePeer(peerID)
}
