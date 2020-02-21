/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog"
	"time"
)

const inTxPeerBufSize = 4

type syncManager struct {
	logger *log.Logger
	actor  p2pcommon.ActorService
	pm     p2pcommon.PeerManager

	blkCache *lru.Cache

	txCache    *lru.Cache
	frontCache map[types.TxID]*incomingTxNotice

	taskChannel   chan smTask
	finishChannel chan struct{}
}

type smTask func()

type incomingTxNotice struct {
	hash     types.TxID
	created  time.Time
	lastSent time.Time
	trial    int
	peers    []types.PeerID
}

func newSyncManager(actor p2pcommon.ActorService, pm p2pcommon.PeerManager, logger *log.Logger) p2pcommon.SyncManager {
	var err error
	sm := &syncManager{actor: actor, pm: pm, logger: logger,
		frontCache:    make(map[types.TxID]*incomingTxNotice),
		taskChannel:   make(chan smTask, 20),
		finishChannel: make(chan struct{}, 1)}

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

func (sm *syncManager) Start() {
	go sm.runManager()
}

func (sm *syncManager) runManager() {
	// set interval of trying to resend getTransaction
	reGetTicker := time.NewTicker(time.Minute)
MANLOOP:
	for {
		select {
		case <-reGetTicker.C:
			sm.refineFrontCache()
		case task := <-sm.taskChannel:
			task()
		case <-sm.finishChannel:
			reGetTicker.Stop()
			break MANLOOP
		}
	}
}

func (sm *syncManager) Stop() {
	close(sm.finishChannel)
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

func (sm *syncManager) RegisterTxNotice(txIDs []types.TxID) {
	sm.taskChannel <- func() {
		for _, txID := range txIDs {
			sm.moveToMPCache(txID)
		}
		sm.logger.Debug().Array("txIDs", types.NewLogTxIDsMarshaller(txIDs, 10)).Msg("syncManager caches txs")
	}
}

func (sm *syncManager) HandleNewTxNotice(peer p2pcommon.RemotePeer, hashes []types.TxID, data *types.NewTransactionsNotice) {
	sm.taskChannel <- func() {
		peerID := peer.ID()
		now := time.Now()
		toGet := make([]message.TXHash, 0, len(hashes))
		for _, txHash := range hashes {
			// check if tx is in cache
			if info, ok := sm.frontCache[txHash]; ok {
				// other peer sent notice already and ready to
				appendPeerID(info, peerID)
				continue
			}
			if ok := sm.txCache.Contains(txHash); ok {
				continue
			}
			// If you want to strict check, query tx to cahinservice. It is skipped since it's so time consuming
			hash := types.HashID(txHash).Bytes()
			toGet = append(toGet, hash)
			sm.frontCache[txHash] = &incomingTxNotice{hash: txHash, created:now, lastSent:now, }
		}
		if len(toGet) == 0 {
			return
		}
		sm.logger.Debug().Int("tx_cnt",len(toGet)).Array("hashes", newLogTXHashesMarshaler(toGet,10)).Msg("syncManager request back unknown tx hashes")
		// create message data
		sm.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
	}
}

//
func (sm *syncManager) refineFrontCache() {
	sm.logger.Debug().Msg("syncManager refine inCache")
	var toSend []*incomingTxNotice
	var deleted []types.TxID
	thresholdTime := time.Now().Add(-time.Minute)
	for txID, info := range sm.frontCache {
		if info.lastSent.Before(thresholdTime) {
			if len(info.peers) > 0 {
				toSend = append(toSend, info)
			} else {
				deleted = append(deleted, txID)
				delete(sm.frontCache, txID)
			}
		}
	}
	sm.logger.Debug().Array("hashes", types.NewLogTxIDsMarshaller(deleted,10)).Msg("syncManager deletes old txids with no responses")

	now := time.Now()
	if len(toSend) == 0 {
		return
	}
	sendMap := make(map[types.PeerID][]message.TXHash)
	for _, info := range toSend {
		peerID := info.peers[0]
		info.peers = info.peers[1:]
		info.lastSent = now
		list, ok := sendMap[peerID]
		if !ok {
			list = make([]message.TXHash,0)
		}
		list = append(list, info.hash[:])
		sendMap[peerID] = list
	}
	for peerID, hashes := range sendMap {
		sm.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Array("hashes", newLogTXHashesMarshaler(hashes,10)).Msg("syncManager retry to get tx to other peers")
		// create message data
		sm.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: hashes})
	}
}


func appendPeerID(info *incomingTxNotice, peerID types.PeerID) {
	info.peers = append(info.peers, peerID)
	if len(info.peers) >= inTxPeerBufSize {
	}
}
func (sm *syncManager) moveToMPCache(txID types.TxID) {
	delete(sm.frontCache, txID)
	sm.txCache.Add(txID, cachePlaceHolder)
}

type logTXHashesMarshaler struct {
	arr []message.TXHash
	limit int
}

func newLogTXHashesMarshaler(bbarray []message.TXHash, limit int) *logTXHashesMarshaler {
	return &logTXHashesMarshaler{arr: bbarray, limit:limit}
}

func (m logTXHashesMarshaler) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(enc.ToString(m.arr[i]))
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(enc.ToString(element))
		}
	}
}