/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	lru "github.com/hashicorp/golang-lru"
)

type syncManager struct {
	logger *log.Logger
	actor  p2pcommon.ActorService
	pm     p2pcommon.PeerManager

	blkCache *lru.Cache
	txCache  *lru.Cache
}

type syncTask struct {
	peer   p2pcommon.RemotePeer
	hashes []types.TxID
	data   *types.NewTransactionsNotice
}

func newSyncManager(actor p2pcommon.ActorService, pm p2pcommon.PeerManager, logger *log.Logger) p2pcommon.SyncManager {
	var err error
	sm := &syncManager{actor: actor, pm: pm, logger: logger}

	sm.blkCache, err = lru.New(DefaultGlobalBlockCacheSize)
	if err != nil {
		panic("Failed to create p2p block cache" + err.Error())
	}
	sm.txCache, err = lru.New(DefaultGlobalTxCacheSize)
	if err != nil {
		panic("Failed to create p2p tx cache " + err.Error())
	}

	return sm
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
		sm.logger.Debug().Str(p2putil.LogBlkHash, enc.ToString(data.BlockHash)).Str(p2putil.LogPeerName, peer.Name()).Msg("new block notice of unknown hash. request back to notifier")
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

func (sm *syncManager) HandleNewTxNotice(peer p2pcommon.RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice) {
	peerID := peer.ID()

	// TODO it will cause problem if getTransaction failed. (i.e. remote peer was sent notice, but not response getTransaction)
	toGet := make([]message.TXHash, 0, len(data.TxHashes))
	for _, hashArr := range hashes {
		ok, _ := sm.txCache.ContainsOrAdd(hashArr, cachePlaceHolder)
		if ok {
			// Kickout duplicated notice log.
			// if sm.logger.IsDebugEnabled() {
			// 	sm.logger.Debug().Str(LogTxHash, enc.ToString(hashArr[:])).Str(LogPeerID, peerID.Pretty()).Msg("Got NewTx notice, but sent already from other peer")
			// }
			// this notice is already sent to chainservice
			continue
		}
		hash := types.HashID(hashArr).Bytes()
		toGet = append(toGet, hash)
	}
	if len(toGet) == 0 {
		// sm.logger.Debug().Str(LogPeerID, peerID.Pretty()).Msg("No new tx found in tx notice")
		return
	}
	sm.logger.Debug().Str("hashes", txHashArrToString(toGet)).Msg("syncManager request back unknown tx hashes")
	// create message data
	sm.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
}

// bytesArrToString converts array of byte array to json array of b58 encoded string.
func txHashArrToString(bbarray []message.TXHash) string {
	return txHashArrToStringWithLimit(bbarray, 10)
}

func txHashArrToStringWithLimit(bbarray []message.TXHash, limit int) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	var arrSize = len(bbarray)
	if limit > arrSize {
		limit = arrSize
	}
	for i := 0; i < limit; i++ {
		hash := bbarray[i]
		buf.WriteByte('"')
		buf.WriteString(enc.ToString([]byte(hash)))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	if arrSize > limit {
		buf.WriteString(fmt.Sprintf(" (and %d more), ", arrSize-limit))
	}
	buf.WriteByte(']')
	return buf.String()
}
