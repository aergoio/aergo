/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package util

import (
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

const (
	// BcNoReorganizing indicates that the blockchain is not under reorganization.
	BcNoReorganizing = iota
	// BcOnReorganizing indicates that the blockchain is under reorganization.
	BcOnReorganizing = iota
)

// BcReorgStatus is a type alias for blockchain reorganization status.
type BcReorgStatus = int32

var logger = log.NewLogger(log.Consensus)

// GetBestBlock returns the current best block from chainservice
func GetBestBlock(hs component.ICompSyncRequester) *types.Block {
	result, err := hs.RequestFuture(message.ChainSvc, &message.GetBestBlock{}, time.Second).Result()
	if err != nil {
		logger.Errorf("failed to get best block info: %v", err.Error())
		return nil
	}
	return result.(message.GetBestBlockRsp).Block
}

// ConnectBlock send an AddBlock request to the chain service.
func ConnectBlock(hs component.ICompSyncRequester, block *types.Block) {
	_, err := hs.RequestFuture(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block}, time.Second).Result()
	if err != nil {
		logger.Errorf("failed to connect block: no=%d, hash=%s, prev=%s",
			block.Header.BlockNo,
			block.ID(),
			block.PrevID())
		return
	}
}

// FetchTXs requests to mempool and returns types.Tx array.
func FetchTXs(hs component.ICompSyncRequester) []*types.Tx {
	//bf.RequestFuture(message.MemPoolSvc, &message.MemPoolGenerateSampleTxs{MaxCount: 3}, time.Second)
	result, err := hs.RequestFuture(message.MemPoolSvc, &message.MemPoolGet{}, time.Second).Result()
	if err != nil {
		logger.Infof("can't fetch transactions from mempool - %v", err)
		return make([]*types.Tx, 0)
	}

	return result.(*message.MemPoolGetRsp).Txs
}

// MaxBlockBodySize returns the maximum block body size.
//
// TODO: This is not an exact size. Let's make it exact!
func MaxBlockBodySize() int {
	return blockchain.DefaultMaxBlockSize - proto.Size(&types.BlockHeader{})
}

// OnReorganizing is a utility function which reports whether *onReorg is set
// to onReorganizing.
func OnReorganizing(onReorg *int32) bool {
	return atomic.LoadInt32(onReorg) == BcOnReorganizing
}

// SetReorganizing is a utility function which sets *onReorg to onReorganizing.
func SetReorganizing(onReorg *int32) {
	atomic.StoreInt32(onReorg, BcOnReorganizing)
}

// UnsetReorganizing is a utility function which sets *onReorg to
// noReorganizing.
func UnsetReorganizing(onReorg *int32) {
	atomic.StoreInt32(onReorg, BcNoReorganizing)
}
