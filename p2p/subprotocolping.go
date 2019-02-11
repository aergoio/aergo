/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type pingRequestHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*pingRequestHandler)(nil)

type pingResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*pingResponseHandler)(nil)

type goAwayHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*goAwayHandler)(nil)

// newPingReqHandler creates handler for PingRequest
func newPingReqHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *pingRequestHandler {
	ph := &pingRequestHandler{BaseMsgHandler: BaseMsgHandler{protocol: PingRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *pingRequestHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.Ping{})
}

func (ph *pingRequestHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := ph.peer
	pingData := msgBody.(*types.Ping)
	debugLogReceiveMsg(ph.logger, ph.protocol, msg.ID().String(), remotePeer, fmt.Sprintf("blockHash=%s blockNo=%d",enc.ToString(pingData.BestBlockHash),pingData.BestHeight))
	if _, err := types.ParseToBlockID(pingData.GetBestBlockHash()); err != nil {
		ph.logger.Info().Str(LogPeerName,remotePeer.Name()).Msg("ping is old format or wrong")
		return
	}
	remotePeer.updateLastNotice(pingData.BestBlockHash, pingData.BestHeight)

	// generate response message
	ph.logger.Debug().Str(LogPeerName,remotePeer.Name()).Str(LogMsgID, msg.ID().String()).Msg("Sending ping response")
	resp := &types.Pong{}
	remotePeer.sendMessage(remotePeer.MF().newMsgResponseOrder(msg.ID(), PingResponse, resp))
}

// newPingRespHandler creates handler for PingResponse
func newPingRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *pingResponseHandler {
	ph := &pingResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: PingResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *pingResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.Pong{})
}

func (ph *pingResponseHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := ph.peer
	//data := msgBody.(*types.Pong)
	debugLogReceiveMsg(ph.logger, ph.protocol, msg.ID().String(), remotePeer, nil)
	remotePeer.consumeRequest(msg.ID())
}

// newGoAwayHandler creates handler for PingResponse
func newGoAwayHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService) *goAwayHandler {
	ph := &goAwayHandler{BaseMsgHandler: BaseMsgHandler{protocol: GoAway, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *goAwayHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GoAwayNotice{})
}

func (ph *goAwayHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	data := msgBody.(*types.GoAwayNotice)
	debugLogReceiveMsg(ph.logger, ph.protocol, msg.ID().String(), ph.peer, data.Message)

	// TODO: check to remove peer here or not. (the sending peer will disconnect.)
}
