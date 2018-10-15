package dpos

import (
	"container/list"
	"fmt"
	"sort"
	"sync"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
)

type errLibUpdate struct {
	current string
	parent  string
	oldBest string
}

func (e errLibUpdate) Error() string {
	return fmt.Sprintf(
		"current block %v (parent %v) inconsistent with old best %v",
		e.current, e.parent, e.oldBest)
}

// Status manages DPoS-related infomations like LIB.
type Status struct {
	sync.RWMutex
	bestBlock *types.Block
	pls       *pLibStatus
	lib       *blockInfo
}

// NewStatus returns a newly allocated Status.
func NewStatus(confirmsRequired uint16) *Status {
	return &Status{
		pls: newPlibStatus(confirmsRequired),
	}
}

type pLibStatus struct {
	genesisInfo      *blockInfo
	confirmsRequired uint16
	confirms         *list.List
	undo             *list.List
	plib             map[string][]*blockInfo // BP-wise proposed LIB map
}

func newPlibStatus(confirmsRequired uint16) *pLibStatus {
	return &pLibStatus{
		confirmsRequired: confirmsRequired,
		confirms:         list.New(),
		undo:             list.New(),
		plib:             make(map[string][]*blockInfo),
	}
}

func (pls *pLibStatus) init() {
	pls.confirms.Init()
}

func (pls *pLibStatus) addConfirmInfo(block *types.Block) {
	ci := newConfirmInfo(block, pls.confirmsRequired)
	pls.confirms.PushBack(ci)

	bi := ci.blockInfo

	// Initialize an empty pre-LIB map entry with genesis block info.
	if _, exist := pls.plib[ci.BPID]; !exist {
		pls.updatePreLIB(ci.BPID, pls.genesisInfo)
	}

	logger.Debug().Str("BP", ci.BPID).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Msg("new confirm info added")
}

func (pls *pLibStatus) updateStatus() *blockInfo {
	if bpID, bi := pls.getPreLIB(); bi != nil {
		pls.updatePreLIB(bpID, bi)

		return pls.calcLIB()
	}
	return nil
}

func (pls *pLibStatus) updatePreLIB(bpID string, bi *blockInfo) {
	pls.plib[bpID] = append(pls.plib[bpID], bi)
	logger.Debug().Str("BP", bpID).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Msg("proposed LIB map updated")
}

func (pls *pLibStatus) rollbackStatusTo(block *types.Block) error {
	var (
		beg           = pls.confirms.Back()
		end           *list.Element
		confirmLow    = cInfo(beg).BlockNo
		targetHash    = block.ID()
		targetBlockNo = block.BlockNo()
	)

	logger.Debug().
		Uint64("target no", targetBlockNo).Int("undo len", pls.undo.Len()).
		Msg("start LIB status rollback")

	// Remove those associated with the blocks reorganized out.
	removeIf(pls.undo,
		func(e *list.Element) bool {
			return cInfo(e).BlockNo > targetBlockNo
		},
	)

	logger.Debug().
		Uint64("target no", targetBlockNo).
		Int("current undo len", pls.undo.Len()).
		Msg("irrelevent element removed from undo list")

	// Check if block is a valid rollback target.
	for e := beg; e != nil; e = e.Prev() {
		c := cInfo(e)
		if min := c.min(); min < confirmLow {
			confirmLow = min
		}

		if c.BlockHash == targetHash {
			end = e
			break
		}
	}

	// XXX To bypass the compile error. TODO: Remove after the LIB recovery is
	// implemented.
	_ = end
	// XXX Temporarily comment out until the LIB recovery is implemented.
	/*
		if end == nil && block.ID() != pls.genesisInfo.BlockHash {
			return fmt.Errorf("not in the main chain: block hash %v, no %v",
				targetHash, block.BlockNo())
		}
	*/

	// Restore the confirm infos in the rollback range by using the undo list.
	pls.restoreConfirms(confirmLow)
	pls.rollback()

	return nil
}

func (pls *pLibStatus) getPreLIB() (bpID string, bi *blockInfo) {
	var (
		prev   *list.Element
		toUndo = false
		e      = pls.confirms.Back()
		cr     = cInfo(e).ConfirmRange
	)
	bpID = cInfo(e).BPID

	for e != nil && cr > 0 {
		prev = e.Prev()
		cr--

		if !toUndo {
			c := cInfo(e)
			c.confirmsLeft--
			if c.confirmsLeft == 0 {
				// proposed LIB info to return
				bi = c.bInfo()
				toUndo = true
			}
		}

		// Move all the previous elements including the one corresponding to a
		// block to be finalized (c.confirmsLeft == 0). Some of them may be
		// restored later as needed for rollback, while others will be removed
		// if LIB is determined.
		if toUndo {
			pls.moveToUndo(e)
		}

		e = prev
	}

	return
}

func (pls *pLibStatus) restoreConfirms(confirmLow uint64) {
	forEach(pls.undo,
		func(e *list.Element) {
			if cInfo(e).BlockNo >= confirmLow {
				moveElem(e, pls.undo, pls.confirms)
			}
		},
	)
}

func (pls *pLibStatus) rollback() {

	// Reset confirmsLeft & collect counts by which confirmLeft must be
	// decreased.
	decCounts := pls.getDecCounts()

	// Decrease confirmLeft & return the new pre-LIB if exists.
	pls.rebuildConfirms(decCounts)

	// Rollback the pre-LIB map based on the new confirms list. -- During
	// rollback, no new pre-LIBs are created. Only some of the existing pre-LIB
	// map entries may be rollback to the previous one.
	pls.rollbackPreLIBs()

	// Don't need to update LIB since there is no other LIB between the LIB and
	// the branch root (rollback target).
}

func (pls *pLibStatus) getDecCounts() map[uint64]uint16 {
	decCounts := make(map[uint64]uint16)

	forEach(pls.confirms,
		func(e *list.Element) {
			c := cInfo(e)
			c.confirmsLeft = pls.confirmsRequired - 1

			for i := c.min(); i < c.BlockNo; i++ {
				if dec, exist := decCounts[i]; exist {
					decCounts[i] = dec + 1
				} else {
					decCounts[i] = 1
				}
			}
		},
	)

	return decCounts
}

func (pls *pLibStatus) rebuildConfirms(decCounts map[uint64]uint16) {
	var lastUndoElem *list.Element

	forEach(pls.confirms,
		func(e *list.Element) {
			c := cInfo(e)
			if dec, exist := decCounts[c.BlockNo]; exist {
				if c.confirmsLeft < dec {
					logger.Debug().Uint64("block no", c.BlockNo).
						Uint16("confirm left", c.confirmsLeft).Uint16("dec count", dec).
						Msg("dec count higher than confirm left")
					c.confirmsLeft = 0
				} else {
					c.confirmsLeft = c.confirmsLeft - dec
				}

				if c.confirmsLeft == 0 {
					lastUndoElem = e
				}
			}
		},
	)

	if lastUndoElem != nil {
		forEachUntil(pls.confirms, lastUndoElem,
			func(e *list.Element) {
				pls.moveToUndoBack(e)
			},
		)
	}
}

func (pls *pLibStatus) rollbackPreLIBs() {
	forEach(pls.confirms,
		func(e *list.Element) {
			pls.rollbackPreLIB(cInfo(e))
		},
	)
}

func (pls *pLibStatus) rollbackPreLIB(c *confirmInfo) {
	if pLib, exist := pls.plib[c.BPID]; exist {
		purgeBeg := len(pLib)
		if purgeBeg == 0 {
			return
		}

		for i, l := range pLib {
			if l.BlockNo >= c.BlockNo {
				purgeBeg = i
				break
			}
		}
		if purgeBeg < len(pLib) {
			oldLen := len(pLib)
			newEntry := pLib[0:purgeBeg]

			pls.plib[c.BPID] = newEntry

			logger.Debug().
				Str("BPID", c.BPID).Int("old len", oldLen).Int("new len", purgeBeg).
				Msg("rollback pre-LIB entry")
		}
	}
}

func (pls *pLibStatus) moveToUndoBack(e *list.Element) {
	moveElemBack(e, pls.confirms, pls.undo)
}

func moveElemBack(e *list.Element, src *list.List, dst *list.List) {
	src.Remove(e)
	dst.PushBack(e.Value)
}

func (pls *pLibStatus) moveToUndo(e *list.Element) {
	moveElem(e, pls.confirms, pls.undo)
}

func moveElem(e *list.Element, src *list.List, dst *list.List) {
	src.Remove(e)
	dst.PushFront(e.Value)
}

func (pls *pLibStatus) gcUndo(lib *blockInfo) {
	removeIf(pls.undo,
		func(e *list.Element) bool {
			return cInfo(e).BlockNo <= lib.BlockNo
		})
}

type pridicate func(e *list.Element) bool

func removeIf(l *list.List, p pridicate) {
	forEach(l,
		func(e *list.Element) {
			if p(e) {
				l.Remove(e)
			}
		},
	)
}

func forEach(l *list.List, f func(e *list.Element)) {
	e := l.Front()
	for e != nil {
		next := e.Next()
		f(e)
		e = next
	}
}

func forEachUntil(l *list.List, end *list.Element, f func(e *list.Element)) {
	e := l.Front()
	for e != nil {
		next := e.Next()
		f(e)
		if e == end {
			break
		}
		e = next
	}
}

func (c *confirmInfo) bInfo() *blockInfo {
	return c.blockInfo
}

func cInfo(e *list.Element) *confirmInfo {
	return e.Value.(*confirmInfo)
}

func (pls *pLibStatus) calcLIB() *blockInfo {
	if len(pls.plib) == 0 {
		return nil
	}

	libInfos := make([]*blockInfo, 0, len(pls.plib))
	for _, l := range pls.plib {
		if len(l) != 0 {
			libInfos = append(libInfos, l[len(l)-1])
		}
	}

	if len(libInfos) == 0 {
		return nil
	}

	sort.Slice(libInfos, func(i, j int) bool {
		return libInfos[i].BlockNo < libInfos[j].BlockNo
	})

	// TODO: check the correctness of the formula.
	lib := libInfos[(len(libInfos)-1)/3]

	return lib
}

type confirmInfo struct {
	*blockInfo
	BPID         string
	confirmsLeft uint16
}

func newConfirmInfo(block *types.Block, confirmsRequired uint16) *confirmInfo {
	return &confirmInfo{
		BPID:         block.BPID2Str(),
		blockInfo:    newBlockInfo(block),
		confirmsLeft: confirmsRequired,
	}
}

func (c confirmInfo) min() uint64 {
	return c.BlockNo - c.ConfirmRange + 1
}

type blockInfo struct {
	BlockHash    string
	BlockNo      uint64
	ConfirmRange uint64
}

func newBlockInfo(block *types.Block) *blockInfo {
	return &blockInfo{
		BlockHash:    block.ID(),
		BlockNo:      block.BlockNo(),
		ConfirmRange: block.GetHeader().GetConfirms(),
	}
}

// UpdateStatus updates the last irreversible block (LIB).
func (s *Status) UpdateStatus(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	var genesisBlock *types.Block

	if s.pls.genesisInfo == nil {
		if genesisBlock = chain.GetGenesisBlock(); genesisBlock != nil {
			s.pls.genesisInfo = &blockInfo{
				BlockHash: genesisBlock.ID(),
				BlockNo:   genesisBlock.BlockNo(),
			}
		}
	}

	if s.bestBlock == nil {
		if initBlock := chain.GetInitialBestBlock(); initBlock != nil {
			s.bestBlock = initBlock
			// Add manually the initial block info to avoid error. TODO: This must
			// be replaced by a correct LIB status recovery process.
			s.pls.addConfirmInfo(s.bestBlock)
		} else {
			s.bestBlock = genesisBlock
		}
	}

	curBestID := s.bestBlock.ID()
	if curBestID == block.PrevID() {
		s.pls.addConfirmInfo(block)

		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.pls.updateStatus(); lib != nil {
			s.updateLIB(lib)
		}
	} else {
		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("target block no", block.BlockNo()).
			Msg("rollback LIB status")

		// Block reorganized. TODO: update consensus status, correctly.
		if err := s.pls.rollbackStatusTo(block); err != nil {
			logger.Debug().Err(err).Msg("failed to rollback DPoS status")
			panic(err)
		}
	}

	s.bestBlock = block
}

func (s *Status) updateLIB(lib *blockInfo) {
	s.lib = lib
	s.pls.gcUndo(lib)

	logger.Debug().
		Str("block hash", s.lib.BlockHash).
		Uint64("block no", s.lib.BlockNo).
		Int("undo len", s.pls.undo.Len()).
		Msg("last irreversible block (BFT) updated")
}

// NeedReorganization reports whether reorganization is needed or not.
func (s *Status) NeedReorganization(rootNo types.BlockNo) bool {
	return true

	/*
			s.RLock()
			defer s.RUnlock()

			if s.lib == nil {
				logger.Debug().Uint64("branch root no", rootNo).Msg("no LIB")
				return true
			}

			libNo := s.lib.BlockNo

			reorganizable := rootNo >= libNo
			if reorganizable {
				logger.Info().
					Uint64("LIB", libNo).
					Uint64("branch root no", rootNo).
					Msg("not reorganizable - the current main branch has a LIB.")
			}

		return reorganizable
	*/
}

func dumpConfirmInfo(name string, l *list.List) {
	forEach(l,
		func(e *list.Element) {
			logger.Debug().Str("confirm info", spew.Sdump(cInfo(e))).Msg(name)
		},
	)
}
