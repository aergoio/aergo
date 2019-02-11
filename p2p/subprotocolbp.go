/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
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


type blockProducedNoticeHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*blockProducedNoticeHandler)(nil)

// newNewBlockNoticeHandler creates handler for NewBlockNotice
func newBlockProducedNoticeHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *blockProducedNoticeHandler {
	bh := &blockProducedNoticeHandler{BaseMsgHandler: BaseMsgHandler{protocol: BlockProducedNotice, pm: pm, sm: sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockProducedNoticeHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.BlockProducedNotice{})
}

func (bh *blockProducedNoticeHandler) handle(msg p2pcommon.Message, msgBody proto.Message) {
	remotePeer := bh.peer
	data := msgBody.(*types.BlockProducedNotice)
	if data.Block == nil || len(data.Block.Hash) == 0 {
		bh.logger.Info().Str(LogPeerName,remotePeer.Name()).Msg("invalid blockProduced notice. block is null")
		return
	}
	// remove to verbose log
	debugLogReceiveMsg(bh.logger, bh.protocol, msg.ID().String(), remotePeer, log.DoLazyEval(func() string {
		return fmt.Sprintf("bp=%s,blk_no=%d,blk_hash=%s", enc.ToString(data.ProducerID), data.BlockNo, enc.ToString(data.Block.Hash))
	}))

	// lru cache can accept hashable key
	block := data.Block
	if _, err := types.ParseToBlockID(data.GetBlock().GetHash()) ; err != nil {
		// TODO add penelty score
		bh.logger.Info().Str(LogPeerName, remotePeer.Name()).Str("hash", enc.ToString(data.GetBlock().GetHash())).Msg("malformed blockHash")
		return
	}
	// block by blockProduced notice must be new fresh block
	remotePeer.updateLastNotice(data.GetBlock().GetHash(), data.BlockNo)
	bh.sm.HandleBlockProducedNotice(bh.peer, block)
}
