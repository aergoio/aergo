/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/hashicorp/golang-lru"
	"sync"
)

type SyncManager interface {
	// handle notice from bp
	HandleBlockProducedNotice(peer RemotePeer, block *types.Block)
	// handle notice from other node
	HandleNewBlockNotice(peer RemotePeer, data *types.NewBlockNotice)
	HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse)
	HandleNewTxNotice(peer RemotePeer, hashes []TxHash, data *types.NewTransactionsNotice)
}

type syncManager struct {
	logger *log.Logger
	actor  ActorService
	pm     PeerManager

	blkCache *lru.Cache
	txCache  *lru.Cache

	syncLock *sync.Mutex
	syncing  bool
}

func newSyncManager(actor ActorService, pm PeerManager, logger *log.Logger) SyncManager {
	var err error
	sm := &syncManager{actor: actor, pm: pm, logger: logger, syncLock: &sync.Mutex{}}

	sm.blkCache, err = lru.New(DefaultGlobalBlockCacheSize)
	if err != nil {
		panic("Failed to create peermanager " + err.Error())
	}
	sm.txCache, err = lru.New(DefaultGlobalTxCacheSize)
	if err != nil {
		panic("Failed to create peermanager " + err.Error())
	}

	return sm
}

func (sm *syncManager) checkWorkToken() bool {
	sm.syncLock.Lock()
	defer sm.syncLock.Unlock()
	return !sm.syncing
}

func (sm *syncManager) HandleBlockProducedNotice(peer RemotePeer, block *types.Block) {
	hash := MustParseBlkHash(block.GetHash())
	ok, _ := sm.blkCache.ContainsOrAdd(hash, cachePlaceHolder)
	if ok {
		sm.logger.Warn().Str(LogBlkHash, hash.String()).Str(LogPeerID, peer.ID().Pretty()).Msg("Duplacated blockProduced notice")
		return
	}
	sm.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peer.ID(), Block: block, Bstate: nil})

}

func (sm *syncManager) HandleNewBlockNotice(peer RemotePeer, data *types.NewBlockNotice) {
	hash := MustParseBlkHash(data.BlockHash)
	peerID := peer.ID()
	//if !sm.checkWorkToken() {
	//	// just ignore it
	//	//sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlockHash)).Str(LogPeerID, peerID.Pretty()).Msg("Ignoring newBlock notice sync syncManager is busy now.")
	//	return
	//}

	// TODO check if evicted return value is needed.
	ok, _ := sm.blkCache.ContainsOrAdd(hash, cachePlaceHolder)
	if ok {
		// Kickout duplicated notice log.
		// if sm.logger.IsDebugEnabled() {
		// 	sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlkHash)).Str(LogPeerID, peerID.Pretty()).Msg("Got NewBlock notice, but sent already from other peer")
		// }
		// this notice is already sent to chainservice
		return
	}

	// request block info if selfnode does not have block already
	foundBlock, _ := sm.actor.GetChainAccessor().GetBlock(data.BlockHash)
	if foundBlock == nil {
		sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlockHash)).Str(LogPeerID, peerID.Pretty()).Msg("new block notice of unknown hash. request back to notifier")
		sm.actor.SendRequest(message.P2PSvc, &message.GetBlockInfos{ToWhom: peerID,
			Hashes: []message.BlockHash{message.BlockHash(data.BlockHash)}})
	}
}
// HandleGetBlockResponse handle when remote peer send a block information.
// TODO this method will be removed after newer syncer is developed
func (sm *syncManager) HandleGetBlockResponse(peer RemotePeer, msg Message, resp *types.GetBlockResponse) {
	blocks := resp.Blocks
	peerID := peer.ID()

	// The response should have only one block here, since this peer had requested only one block.
	// getblockresponse with bulky blocks is only called in newsyncer since aergosvr 0.9.9 , which is handled by other receiver and not come to this code.
	// if bulky hashes on this condition block, it is probably sync timeout or bug.
	if len(blocks) != 1 {
		return
	}
	block := blocks[0]
	sm.actor.SendRequest(message.ChainSvc, &message.AddBlock{PeerID: peerID, Block: block, Bstate: nil})
}

func (sm *syncManager) HandleNewTxNotice(peer RemotePeer, hashArrs []TxHash, data *types.NewTransactionsNotice) {
	peerID := peer.ID()

	// TODO it will cause problem if getTransaction failed. (i.e. remote peer was sent notice, but not response getTransaction)
	toGet := make([]message.TXHash, 0, len(data.TxHashes))
	for _, hashArr := range hashArrs {
		ok, _ := sm.txCache.ContainsOrAdd(hashArr, cachePlaceHolder)
		if ok {
			// Kickout duplicated notice log.
			// if sm.logger.IsDebugEnabled() {
			// 	sm.logger.Debug().Str(LogTxHash, enc.ToString(hashArr[:])).Str(LogPeerID, peerID.Pretty()).Msg("Got NewTx notice, but sent already from other peer")
			// }
			// this notice is already sent to chainservice
			continue
		}
		hash := make([]byte, txhashLen)
		copy(hash, hashArr[:])
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

func blockHashArrToString(bbarray []message.BlockHash) string {
	return blockHashArrToStringWithLimit(bbarray, 10)
}

func blockHashArrToStringWithLimit(bbarray []message.BlockHash, limit int ) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	var arrSize = len(bbarray)
	if limit > arrSize {
		limit = arrSize
	}
	for i :=0; i < limit; i++ {
		hash := bbarray[i]
		buf.WriteByte('"')
		buf.WriteString(enc.ToString([]byte(hash)))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	if arrSize > limit {
		buf.WriteString(fmt.Sprintf(" (and %d more), ",  arrSize - limit))
	}
	buf.WriteByte(']')
	return buf.String()
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

// bytesArrToString converts array of byte array to json array of b58 encoded string.
func P2PTxHashArrToString(bbarray []TxHash) string {
	return P2PTxHashArrToStringWithLimit(bbarray, 10)
}
func P2PTxHashArrToStringWithLimit(bbarray []TxHash, limit int) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	var arrSize = len(bbarray)
	if limit > arrSize {
		limit = arrSize
	}
	for i := 0; i < limit; i++ {
		hash := bbarray[i]
		buf.WriteByte('"')
		buf.WriteString(enc.ToString(hash[:]))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	if arrSize > limit {
		buf.WriteString(fmt.Sprintf(" (and %d more), ", arrSize-limit))
	}
	buf.WriteByte(']')
	return buf.String()
}
