/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type getClusterRequestHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getClusterRequestHandler)(nil)

type getClusterResponseHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*getClusterResponseHandler)(nil)

// NewGetClusterReqHandler creates handler for PingRequest
func NewGetClusterReqHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getClusterRequestHandler {
	ph := &getClusterRequestHandler{BaseMsgHandler{protocol: GetClusterRequest, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *getClusterRequestHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.AddressesRequest{})
}

func (ph *getClusterRequestHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	//peerID := ph.peer.ID()
	remotePeer := ph.peer
	data := msgBody.(*types.GetClusterInfoRequest)
	p2putil.DebugLogReceiveMsg(ph.logger, ph.protocol, msg.ID().String(), remotePeer, data.String())

	// TODO handle
	resp := &types.GetClusterInfoResponse{Status:types.ResultStatus_UNIMPLEMENTED}
	remotePeer.SendMessage(remotePeer.MF().NewMsgResponseOrder(msg.ID(), GetClusterResponse, resp))
}


// NewGetClusterRespHandler creates handler for PingRequest
func NewGetClusterRespHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService) *getClusterResponseHandler {
	ph := &getClusterResponseHandler{BaseMsgHandler{protocol: GetClusterResponse, pm: pm, peer: peer, actor: actor, logger: logger}}
	return ph
}

func (ph *getClusterResponseHandler) ParsePayload(rawbytes []byte) (proto.Message, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.AddressesResponse{})
}

func (ph *getClusterResponseHandler) Handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := ph.peer
	data := msgBody.(*types.GetClusterInfoResponse)
	p2putil.DebugLogReceiveResponseMsg(ph.logger, ph.protocol, msg.ID().String(), msg.OriginalID().String(), remotePeer, data.Status.String())

	remotePeer.ConsumeRequest(msg.OriginalID())
	// TODO do handle response

}
