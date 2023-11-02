/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	lru "github.com/hashicorp/golang-lru"
)

const inTxPeerBufSize = 4

type syncManager struct {
	logger *log.Logger
	actor  p2pcommon.ActorService
	pm     p2pcommon.PeerManager

	tm *syncTxManager

	blkCache *lru.Cache
}

func newSyncManager(actor p2pcommon.ActorService, pm p2pcommon.PeerManager, logger *log.Logger) p2pcommon.SyncManager {
	var err error
	sm := &syncManager{actor: actor, pm: pm, logger: logger}
	sm.tm = newTxSyncManager(sm, actor, pm, logger)

	sm.blkCache, err = lru.New(DefaultGlobalBlockCacheSize)
	if err != nil {
		panic("Failed to create p2p block cache" + err.Error())
	}

	return sm
}

func (sm *syncManager) Start() {
	sm.tm.Start()
}

func (sm *syncManager) Stop() {
	sm.tm.Stop()
}

func (sm *syncManager) HandleBlockProducedNotice(peer p2pcommon.RemotePeer, block *types.Block) {
	hash := types.MustParseBlockID(block.GetHash())
	ok, _ := sm.blkCache.ContainsOrAdd(hash, syncManagerChanSize)
	if ok {
		sm.logger.Warn().Str(p2putil.LogBlkHash, hash.String()).Str(p2putil.LogPeerName, peer.Name()).Msg("Duplicated blockProduced notice")
		return
	}
	// check if block size is over the limit
	if block.Size() > int(chain.MaxBlockSize()) {
		sm.logger.Info().Str(p2putil.LogPeerName, peer.Name()).Str(p2putil.LogBlkHash, block.BlockID().String()).Int("size", block.Size()).Msg("invalid blockProduced notice. block size exceed limit")
		return
	}

	sm.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peer.ID(), Block: block, Bstate: nil})
}

func (sm *syncManager) HandleNewBlockNotice(peer p2pcommon.RemotePeer, data *types.NewBlockNotice) {
	hash := types.MustParseBlockID(data.BlockHash)
	peerID := peer.ID()
	//if !sm.checkWorkToken() {
	//	// just ignore it
	//	//sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlockHash)).Str(LogPeerID, peerID.Pretty()).Msg("Ignoring newBlock notice sync syncManager is busy now.")
	//	return
	//}

	ok, _ := sm.blkCache.ContainsOrAdd(hash, cachePlaceHolder)
	if ok {
		// Kick out duplicated notice log.
		// if sm.logger.IsDebugEnabled() {
		// 	sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlkHash)).Str(LogPeerID, peerID.Pretty()).Msg("Got NewBlock notice, but sent already from other peer")
		// }
		// this notice is already sent to chainservice
		return
	}

	// request block info if selfnode does not have block already
	foundBlock, _ := sm.actor.GetChainAccessor().GetBlock(data.BlockHash)
	if foundBlock == nil {
		sm.logger.Debug().Stringer(p2putil.LogBlkHash, types.LogBase58(data.BlockHash)).Str(p2putil.LogPeerName, peer.Name()).Msg("new block notice of unknown hash. request back to notifier")
		sm.actor.SendRequest(message.P2PSvc, &message.GetBlockInfos{ToWhom: peerID,
			Hashes: []message.BlockHash{message.BlockHash(data.BlockHash)}})
	}
}

// HandleGetBlockResponse handle when remote peer send a block information.
// TODO this method will be removed after newer syncer is developed
func (sm *syncManager) HandleGetBlockResponse(peer p2pcommon.RemotePeer, msg p2pcommon.Message, resp *types.GetBlockResponse) {
	blocks := resp.Blocks
	peerID := peer.ID()

	// The response should have only one block here, since this peer had requested only one block.
	// getBlockResponse with bulky blocks is only called in newsyncer since aergosvr 0.9.9 , which is handled by other receiver and not come to this code.
	// if bulky hashes on this condition block, it is probably sync timeout or bug.
	if len(blocks) != 1 {
		return
	}
	block := blocks[0]
	// check if block size is over the limit
	if block.Size() > int(chain.MaxBlockSize()) {
		sm.logger.Info().Str(p2putil.LogPeerName, peer.Name()).Str(p2putil.LogBlkHash, block.BlockID().String()).Int("size", block.Size()).Msg("cancel to add block. block size exceed limit")
		return
	}

	sm.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block, Bstate: nil})
}

func (sm *syncManager) RegisterTxNotice(txs []*types.Tx) {
	sm.tm.registerTxNotice(txs)
}

func (sm *syncManager) HandleNewTxNotice(peer p2pcommon.RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice) {
	sm.tm.HandleNewTxNotice(peer, hashes, data)
}

func (sm *syncManager) HandleGetTxReq(peer p2pcommon.RemotePeer, msgID p2pcommon.MsgID, data *types.GetTransactionsRequest) error {
	return sm.tm.HandleGetTxReq(peer, msgID, data)
}

func (sm *syncManager) RetryGetTx(peer p2pcommon.RemotePeer, hashes [][]byte) {
	sm.tm.retryGetTx(peer.ID(), hashes)
}

func (sm *syncManager) Summary() map[string]interface{} {
	type sizes struct {
		fcSize, qSize int
	}
	var retChan = make(chan sizes, 1)
	sm.tm.taskChannel <- func() {
		retChan <- sizes{fcSize: len(sm.tm.frontCache), qSize: sm.tm.toNoticeIdQueue.Len()}
	}

	txMap := make(map[string]interface{})
	select {
	case s := <-retChan:
		txMap["queryQueue"] = s.qSize
		txMap["frontCache"] = s.fcSize
	case <-time.NewTimer(time.Millisecond << 4).C:
		// timeout
	}
	// There can be a little error
	sum := make(map[string]interface{})
	sum["blocks"] = sm.blkCache.Len()
	txMap["backCache"] = sm.tm.txCache.Len()
	sum["transactions"] = txMap
	return sum
}
