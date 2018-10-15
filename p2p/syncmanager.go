/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/hashicorp/golang-lru"
	"reflect"
)

type SyncManager interface {
	HandleNewBlockNotice(peer RemotePeer, hash BlockHash, data *types.NewBlockNotice)
	HandleNewTxNotice(peer RemotePeer, hashes []TxHash, data *types.NewTransactionsNotice)
}

type syncManager struct {
	logger *log.Logger
	actor ActorService
	pm PeerManager

	blkCache *lru.Cache
	txCache  *lru.Cache
}

func newSyncManager(actor ActorService, pm PeerManager, logger *log.Logger) SyncManager {
	var err error
	sm := &syncManager{actor:actor, pm:pm, logger:logger}

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

func (sm *syncManager) HandleNewBlockNotice(peer RemotePeer, hashArr BlockHash, data *types.NewBlockNotice) {
	peerID := peer.ID()

	// TODO check if evicted return value is needed.
	ok, _ := sm.blkCache.ContainsOrAdd(hashArr, cachePlaceHolder)
	if ok {
		// Kickout duplicated notice log.
		// if sm.logger.IsDebugEnabled() {
		// 	sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlockHash)).Str(LogPeerID, peerID.Pretty()).Msg("Got NewBlock notice, but sent already from other peer")
		// }
		// this notice is already sent to chainservice
		return
	}

	// request block info if selfnode does not have block already
	rawResp, err := sm.actor.CallRequest(message.ChainSvc, &message.GetBlock{BlockHash: message.BlockHash(data.BlockHash)})
	if err != nil {
		sm.logger.Warn().Err(err).Msg("actor return error on getblock")
		return
	}
	resp, ok := rawResp.(message.GetBlockRsp)
	if !ok {
		sm.logger.Warn().Str("expected", "message.GetBlockRsp").Str("actual", reflect.TypeOf(rawResp).Name()).Msg("chainservice returned unexpected type")
		return
	}
	if resp.Err != nil {
		sm.logger.Debug().Str(LogBlkHash, enc.ToString(data.BlockHash)).Str(LogPeerID, peerID.Pretty()).Msg("chainservice responded that block not found. request back to notifier")
		sm.actor.SendRequest(message.P2PSvc, &message.GetBlockInfos{ToWhom: peerID,
			Hashes: []message.BlockHash{message.BlockHash(data.BlockHash)}})
	}

}

func (sm *syncManager) HandleNewTxNotice(peer RemotePeer, hashArrs []TxHash, data *types.NewTransactionsNotice) {
	peerID := peer.ID()

	sm.logger.Debug().Str("hashes", P2PTxHashArrToString(hashArrs)).Msg("inputs")

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
	sm.logger.Debug().Str("hashes", txHashArrToString(toGet)).Msg("toGet is")
	// create message data
	sm.actor.SendRequest(message.P2PSvc, &message.GetTransactions{ToWhom: peerID, Hashes: toGet})
}

// bytesArrToString converts array of byte array to json array of b58 encoded string.
func txHashArrToString(bbarray []message.TXHash) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, hash := range bbarray {
		buf.WriteByte('"')
		buf.WriteString(enc.ToString([]byte(hash)))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	buf.WriteByte(']')
	return buf.String()
}

// bytesArrToString converts array of byte array to json array of b58 encoded string.
func P2PTxHashArrToString(bbarray []TxHash) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, hash := range bbarray {
		buf.WriteByte('"')
		buf.WriteString(enc.ToString(hash[:]))
		buf.WriteByte('"')
		buf.WriteByte(',')
	}
	buf.WriteByte(']')
	return buf.String()
}