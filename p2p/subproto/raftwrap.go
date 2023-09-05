/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"context"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/p2p/raftsupport"
	"github.com/aergoio/etcd/raft/raftpb"
)

// receive message, decode payload to raftpb.Message and toss it to raft
type raftWrapperHandler struct {
	BaseMsgHandler

	consAcc consensus.ConsensusAccessor
}

var _ p2pcommon.MessageHandler = (*raftWrapperHandler)(nil)

// NewGetClusterReqHandler creates handler for PingRequest
func NewRaftWrapperHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, consAcc consensus.ConsensusAccessor) *raftWrapperHandler {
	ph := &raftWrapperHandler{
		BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.RaftWrapperMessage, pm: pm, peer: peer, actor: actor, logger: logger},
		consAcc:        consAcc,
	}
	return ph
}

func (ph *raftWrapperHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &raftpb.Message{})
}

func (ph *raftWrapperHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := ph.peer
	data := msgBody.(*raftpb.Message)
	if ph.logger.IsDebugEnabled() &&
		(data.Type != raftpb.MsgHeartbeatResp && data.Type != raftpb.MsgHeartbeat) {
		DebugLogRaftWrapMsg(ph.logger, remotePeer, msg.ID(), data)
	}

	// toss data to raft module
	if err := ph.consAcc.RaftAccessor().Process(context.TODO(), remotePeer.ID(), *data); err != nil {
		ph.logger.Debug().Str(p2putil.LogPeerName, remotePeer.Name()).Err(err).Msg("error while processing raft message ")
	}
}

func (ph *raftWrapperHandler) PostHandle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data, ok := msgBody.(*raftpb.Message)
	if ok && (data.Type != raftpb.MsgHeartbeatResp && data.Type != raftpb.MsgHeartbeat) {
		// TODO change to show more meaningful information of raft message
		ph.BaseMsgHandler.PostHandle(msg, msgBody)
	}
}

func DebugLogRaftWrapMsg(logger *log.Logger, peer p2pcommon.RemotePeer, msgID p2pcommon.MsgID, body *raftpb.Message) {
	logger.Debug().Str(p2putil.LogMsgID, msgID.String()).Str("from_peer", peer.Name()).Object("raftMsg", raftsupport.RaftMsgMarshaller{Message: body}).Msg("Received raft message")
}
