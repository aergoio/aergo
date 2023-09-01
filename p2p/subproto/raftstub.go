/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

// raftBPNoticeDiscardHandler silently discard blk notice. It is for raft block producer, since raft BP receive notice from raft HTTPS
type raftBPNoticeDiscardHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*raftBPNoticeDiscardHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewBPNoticeDiscardHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) p2pcommon.MessageHandler {
	bh := &raftBPNoticeDiscardHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.BlockProducedNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *raftBPNoticeDiscardHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.BlockProducedNotice{})
}

func (bh *raftBPNoticeDiscardHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.BlockProducedNotice)
	if data.GetBlock() == nil || len(data.GetBlock().Hash) == 0 {
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid blockProduced notice. block is null")
		return
	}
	// just update last status
	if blockID, err := types.ParseToBlockID(data.GetBlock().Hash); err != nil {
		bh.logger.Info().Err(err).Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid block hash")
		return
	} else {
		remotePeer.UpdateLastNotice(blockID, data.BlockNo)
	}
}

// raftBPNoticeDiscardHandler silently discard blk notice. It is for raft block producer, since raft BP receive notice from raft HTTPS
type raftNewBlkNoticeDiscardHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*raftNewBlkNoticeDiscardHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewBlkNoticeDiscardHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) p2pcommon.MessageHandler {
	bh := &raftNewBlkNoticeDiscardHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.NewBlockNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *raftNewBlkNoticeDiscardHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.NewBlockNotice{})
}

func (bh *raftNewBlkNoticeDiscardHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.NewBlockNotice)

	if blockID, err := types.ParseToBlockID(data.BlockHash); err != nil {
		// TODO Add penalty score and break
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.BlockHash)).Msg("malformed blockHash")
		return
	} else {
		// just update last status
		remotePeer.UpdateLastNotice(blockID, data.BlockNo)
	}
}
