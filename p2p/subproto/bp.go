/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
)

type blockProducedNoticeHandler struct {
	BaseMsgHandler
}

var _ p2pcommon.MessageHandler = (*blockProducedNoticeHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func NewBlockProducedNoticeHandler(pm p2pcommon.PeerManager, peer p2pcommon.RemotePeer, logger *log.Logger, actor p2pcommon.ActorService, sm p2pcommon.SyncManager) *blockProducedNoticeHandler {
	bh := &blockProducedNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: p2pcommon.BlockProducedNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockProducedNoticeHandler) ParsePayload(rawbytes []byte) (p2pcommon.MessageBody, error) {
	return p2putil.UnmarshalAndReturn(rawbytes, &types.BlockProducedNotice{})
}

func (bh *blockProducedNoticeHandler) Handle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	remotePeer := bh.peer
	data := msgBody.(*types.BlockProducedNotice)
	if data.Block == nil || len(data.Block.Hash) == 0 {
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Msg("invalid blockProduced notice. block is null")
		return
	}
	// remove to verbose log
	p2putil.DebugLogReceive(bh.logger, bh.protocol, msg.ID().String(), remotePeer,data)
	
	// lru cache can accept hashable key
	block := data.Block
	if blockID, err := types.ParseToBlockID(data.GetBlock().GetHash()); err != nil {
		// TODO add penalty score
		bh.logger.Info().Str(p2putil.LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.GetBlock().GetHash())).Msg("malformed blockHash")
		return
	} else {
		// block by blockProduced notice must be new fresh block
		remotePeer.UpdateLastNotice(blockID, data.BlockNo)
		bh.sm.HandleBlockProducedNotice(bh.peer, block)
	}
}
