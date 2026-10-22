/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type pingRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*pingRequestHandler)(nil)

type pingResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*pingResponseHandler)(nil)

type goAwayHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*goAwayHandler)(nil)

// newPingReqHandler creates handler for PingRequest
func NewPingReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *pingRequestHandler {
	ph := &pingRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.PingRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *pingRequestHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.Ping{})
}

func (ph *pingRequestHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := ph.peer
	pingData := msgBody.(*types.Ping)
	p2putil.DebugLogReceive(ph.logger, ph.protocol, msg.ID().String(), remotePeer, pingData)
	if blockID, err := types.ParseToBlockID(pingData.GetBestBlockHash()); err != nil {
		ph.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("ping is old format or wrong")
		return
	} else {
		remotePeer.UpdateLastNotice(blockID, pingData.BestHeight)
	}

	// generate response message
	ph.logger.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Str(p2putil.LogMsgID, msg.ID().String()).Msg("Sending ping response")
	resp := &types.Pong{}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), p2pcommon.PingResponse, resp))
}

// newPingRespHandler creates handler for PingResponse
func NewPingRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *pingResponseHandler {
	ph := &pingResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.PingResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *pingResponseHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.Pong{})
}

func (ph *pingResponseHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := ph.peer
	//data := msgBody.(*types.Pong)
	p2putil.DebugLogReceive(ph.logger, ph.protocol, msg.ID().String(), remotePeer, nil)
	remotePeer.ConsumeRequest(msg.ID())
}

// newGoAwayHandler creates handler for PingResponse
func NewGoAwayHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *goAwayHandler {
	ph := &goAwayHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.GoAway, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *goAwayHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.GoAwayNotice{})
}

func (ph *goAwayHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data := msgBody.(*types.GoAwayNotice)
	p2putil.DebugLogReceive(ph.logger, ph.protocol, msg.ID().String(), ph.peer, data)

	// TODO: check to remove peer here or not. (the sending peer will disconnect.)
}
