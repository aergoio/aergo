/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/aergoio/aergo/types"
)

type OrphanBlock struct {
	block      *types.Block
	expiretime time.Time
}

type OrphanPool struct {
	sync.RWMutex
	cache  map[types.BlockKey]*OrphanBlock
	maxCnt int
	curCnt int
}

func NewOrphanPool() *OrphanPool {
	return &OrphanPool{
		cache:  map[types.BlockKey]*OrphanBlock{},
		maxCnt: 1000,
		curCnt: 0,
	}
}

// add Orphan into the orphan cache pool
func (op *OrphanPool) addOrphan(block *types.Block) error {
	key := types.ToBlockKey(block.Header.PrevBlockHash)
	cachedblock, exists := op.cache[key]
	if exists {
		logger.Debug().Bytes("hash", block.GetHash()).Bytes("cashed", cachedblock.block.GetHash()).Msg("already exist")
		return fmt.Errorf("orphan block already exist")
	}

	if op.maxCnt == op.curCnt {
		logger.Debug().Msg("orphan block pool is full")
		// replace one
		op.removeOldest()
	}
	op.cache[key] = &OrphanBlock{
		block:      block,
		expiretime: time.Now().Add(time.Hour),
	}
	op.curCnt++
	logger.Debug().Bytes("hash", block.GetHash()).Msg("add Orphan Block")
	return nil
}

// get the BlockKey of Root Block of Orphan branch
func (op *OrphanPool) getRoot(block *types.Block) types.BlockKey {
	orphanRoot := types.ToBlockKey(block.Header.PrevBlockHash)
	prevKey := orphanRoot
	for {
		orphan, exists := op.cache[prevKey]
		if !exists {
			break
		}
		orphanRoot = prevKey
		prevKey = types.ToBlockKey(orphan.block.Header.PrevBlockHash)
	}

	return orphanRoot
}

// remove oldest block, but also remove expired
func (op *OrphanPool) removeOldest() {
	// remove all expired
	var oldest *OrphanBlock
	for key, orphan := range op.cache {
		if time.Now().After(orphan.expiretime) {
			logger.Debug().Str("hash", key.String()).Msg("orphan block removed(expired)")
			op.removeOrphan(key)
		}

		// choose at least one victim
		if oldest == nil || orphan.expiretime.Before(oldest.expiretime) {
			oldest = orphan
		}
	}

	// remove oldest one
	if op.curCnt == op.maxCnt {
		key := types.ToBlockKey(oldest.block.Header.PrevBlockHash)
		logger.Debug().Str("hash", key.String()).Msg("orphan block removed(oldest)")
		op.removeOrphan(key)
	}
}

// remove one single element by key (must succeed)
func (op *OrphanPool) removeOrphan(key types.BlockKey) {
	delete(op.cache, key)
	op.curCnt--
}
