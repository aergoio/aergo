/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/rs/zerolog"
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

	// TODO toss data to raft module
}

func (ph *raftWrapperHandler) PostHandle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	data, ok := msgBody.(*raftpb.Message)
	if ok && (data.Type != raftpb.MsgHeartbeatResp && data.Type != raftpb.MsgHeartbeat) {
		// TODO change to show more meaningful information of raft message
		ph.BaseMsgHandler.PostHandle(msg, msgBody)
	}
}



func DebugLogRaftWrapMsg(logger *log.Logger, peer p2pcommon.RemotePeer, msgID p2pcommon.MsgID, body *raftpb.Message) {
	logger.Debug().Str(p2putil.LogMsgID, msgID.String()).Str("from_peer", peer.Name()).Object("raftMsg", RaftMsgMarshaller{body}).Msg("Received raft message")
}

type RaftMsgMarshaller struct {
	*raftpb.Message
}

func (m RaftMsgMarshaller) MarshalZerologObject(e *zerolog.Event) {
	e.Str("type", m.Type.String()).Uint64("term", m.Term).Uint64("index", m.Index)
}


