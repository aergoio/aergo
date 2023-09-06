/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"sync"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/hashicorp/golang-lru/simplelru"
)

var (
	DfltOrphanPoolSize = 100

	ErrRemoveOldestOrphan = errors.New("failed to remove oldest orphan block")
	ErrNotExistOrphanLRU  = errors.New("given orphan doesn't exist in lru")
)

type OrphanBlock struct {
	*types.Block
}

type OrphanPool struct {
	sync.RWMutex
	cache map[types.BlockID]*OrphanBlock
	lru   *simplelru.LRU

	maxCnt int
	curCnt int
}

func NewOrphanPool(size int) *OrphanPool {
	lru, err := simplelru.NewLRU(DfltOrphanPoolSize, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init lru")
		return nil
	}

	return &OrphanPool{
		cache:  map[types.BlockID]*OrphanBlock{},
		lru:    lru,
		maxCnt: size,
		curCnt: 0,
	}
}

// add Orphan into the orphan cache pool
func (op *OrphanPool) addOrphan(block *types.Block) error {
	logger.Warn().Str("prev", enc.ToString(block.GetHeader().GetPrevBlockHash())).Msg("add orphan Block")

	id := types.ToBlockID(block.Header.PrevBlockHash)
	cachedblock, exists := op.cache[id]
	if exists {
		logger.Debug().Str("hash", block.ID()).
			Str("cached", cachedblock.ID()).Msg("already exist")
		return nil
	}

	if op.isFull() {
		logger.Debug().Msg("orphan block pool is full")
		// replace one
		if err := op.removeOldest(); err != nil {
			return err
		}
	}

	orpEntry := &OrphanBlock{Block: block}

	op.cache[id] = orpEntry
	op.lru.Add(id, orpEntry)
	op.curCnt++

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
		prevID = types.ToBlockID(orphan.Header.PrevBlockHash)
	}

	return orphanRoot
}

func (op *OrphanPool) isFull() bool {
	return op.maxCnt == op.curCnt
}

// remove oldest block, but also remove expired
func (op *OrphanPool) removeOldest() error {
	var (
		id types.BlockID
	)

	if !op.isFull() {
		return nil
	}

	key, _, ok := op.lru.GetOldest()
	if !ok {
		return ErrRemoveOldestOrphan
	}

	id = key.(types.BlockID)
	if err := op.removeOrphan(id); err != nil {
		return err
	}

	logger.Debug().Str("hash", id.String()).Msg("orphan block removed(oldest)")

	return nil
}

// remove one single element by id (must succeed)
func (op *OrphanPool) removeOrphan(id types.BlockID) error {
	op.curCnt--
	delete(op.cache, id)
	if exist := op.lru.Remove(id); !exist {
		return ErrNotExistOrphanLRU
	}
	return nil
}

func (op *OrphanPool) getOrphan(hash []byte) *types.Block {
	prevID := types.ToBlockID(hash)

	orphan, exists := op.cache[prevID]
	if !exists {
		return nil
	} else {
		return orphan.Block
	}
}
