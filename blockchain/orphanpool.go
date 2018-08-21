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
	cache  map[types.BlockID]*OrphanBlock
	maxCnt int
	curCnt int
}

func NewOrphanPool() *OrphanPool {
	return &OrphanPool{
		cache:  map[types.BlockID]*OrphanBlock{},
		maxCnt: 1000,
		curCnt: 0,
	}
}

// add Orphan into the orphan cache pool
func (op *OrphanPool) addOrphan(block *types.Block) error {
	id := types.ToBlockID(block.Header.PrevBlockHash)
	cachedblock, exists := op.cache[id]
	if exists {
		logger.Debug().Str("hash", block.ID()).
			Str("cached", cachedblock.block.ID()).Msg("already exist")
		return fmt.Errorf("orphan block already exist")
	}

	if op.maxCnt == op.curCnt {
		logger.Debug().Msg("orphan block pool is full")
		// replace one
		op.removeOldest()
	}
	op.cache[id] = &OrphanBlock{
		block:      block,
		expiretime: time.Now().Add(time.Hour),
	}
	op.curCnt++
	logger.Debug().Str("hash", block.ID()).Msg("add Orphan Block")
	return nil
}

// get the BlockID of Root Block of Orphan branch
func (op *OrphanPool) getRoot(block *types.Block) types.BlockID {
	orphanRoot := types.ToBlockID(block.Header.PrevBlockHash)
	prevID := orphanRoot
	for {
		orphan, exists := op.cache[prevID]
		if !exists {
			break
		}
		orphanRoot = prevID
		prevID = types.ToBlockID(orphan.block.Header.PrevBlockHash)
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
		id := types.ToBlockID(oldest.block.Header.PrevBlockHash)
		logger.Debug().Str("hash", id.String()).Msg("orphan block removed(oldest)")
		op.removeOrphan(id)
	}
}

// remove one single element by id (must succeed)
func (op *OrphanPool) removeOrphan(id types.BlockID) {
	delete(op.cache, id)
	op.curCnt--
}
