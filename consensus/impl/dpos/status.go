package dpos

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
)

type preLIB = map[string][]*blockInfo

var (
	statusKeyLIB    = []byte("dposStatus.LIB")
	statusKeyPreLIB = []byte("dposStatus.PreLIB")
	bootState       *bootingStatus
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
	bestBlock   *types.Block
	pls         *pLibStatus
	lib         *blockInfo
	initialized bool
}

type bootingStatus struct {
	plib    preLIB
	lib     *blockInfo
	best    *types.Block
	genesis *types.Block
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
	plib             preLIB // BP-wise proposed LIB map
}

func newPlibStatus(confirmsRequired uint16) *pLibStatus {
	return &pLibStatus{
		confirmsRequired: confirmsRequired,
		confirms:         list.New(),
		undo:             list.New(),
		plib:             make(preLIB),
	}
}

func (pls *pLibStatus) init() {
	pls.confirms.Init()
}

func (pls *pLibStatus) addConfirmInfo(block *types.Block) {
	// Genesis block must not be added.
	if block.BlockNo() == 0 {
		return
	}

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

func (pls *pLibStatus) update() *blockInfo {
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
			pls.confirms.Remove(e)
			pls.addToUndo(e)
		}

		e = prev
	}

	return
}

func (pls *pLibStatus) restoreConfirms(confirmLow uint64) {
	// Elements of pls.undo are in asceding of its block no. The confirms list
	// elements must be also in ascending order. This is why confirms list is
	// reversely traversed and each removed element is pushed into the front.
	forEachReverse(pls.undo,
		func(e *list.Element) {
			if cInfo(e).BlockNo >= confirmLow {
				moveElemToFront(e, pls.undo, pls.confirms)
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
				pls.confirms.Remove(e)
				pls.addToUndo(e)
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

func (pls *pLibStatus) addToUndo(newElem *list.Element) {
	var mark *list.Element
	ci := cInfo(newElem)

	// Maintain elements in ascending order of block no.
	for e := pls.undo.Front(); e != nil; e = e.Next() {
		if ci.BlockNo < cInfo(e).BlockNo {
			mark = e
			break

		}
	}

	if mark != nil {
		pls.undo.InsertBefore(newElem.Value, mark)
	} else {
		pls.undo.PushBack(newElem.Value)
	}

	/*
		if mark != nil {
			dumpConfirmInfo(
				fmt.Sprintf("XXX elem: %v, mark: %v XXX (len=%v)",
					cInfo(newElem).BlockNo,
					cInfo(mark).BlockNo,
					pls.undo.Len(),
				), pls.undo)
		} else {
			dumpConfirmInfo(
				fmt.Sprintf("XXX elem to tail: %v (len=%v)",
					cInfo(newElem).BlockNo, pls.undo.Len()),
				pls.undo)
		}
	*/
}

func moveElemToFront(e *list.Element, src *list.List, dst *list.List) {
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

func forEachReverse(l *list.List, f func(e *list.Element)) {
	e := l.Back()
	for e != nil {
		prev := e.Prev()
		f(e)
		e = prev
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

// Init recovers the last DPoS status including pre-LIB map and confirms
// list between LIB and the best block.
func (s *Status) Init(genesis, best *types.Block, get func([]byte) []byte,
	getBlock func(types.BlockNo) *types.Block) {

	bootState = &bootingStatus{
		plib:    make(preLIB),
		lib:     &blockInfo{},
		best:    best,
		genesis: genesis,
	}

	bootState.loadLIB(get)
}

func (bs *bootingStatus) loadLIB(get func([]byte) []byte) {
	decodeStatus := func(key []byte, dst interface{}) error {
		value := get(key)
		if len(value) == 0 {
			return fmt.Errorf("LIB status not found: key = %v", string(key))
		}

		err := decode(bytes.NewBuffer(value), dst)
		if err != nil {
			logger.Debug().Err(err).Str("key", string(key)).
				Msg("failed to decode DPoS status")
			panic(err)
		}
		return nil
	}

	if err := decodeStatus(statusKeyLIB, bs.lib); err == nil {
		logger.Debug().Uint64("block no", bs.lib.BlockNo).
			Str("block hash", bs.lib.BlockHash).Msg("LIB loaded from DB")
	}

	if err := decodeStatus(statusKeyPreLIB, &bs.plib); err == nil {
		logger.Debug().Int("len", len(bs.plib)).Msg("pre-LIB loaded from DB")
		for id, p := range bs.plib {
			logger.Debug().
				Str("BPID", id).Str("block hash", p[len(p)-1].BlockHash).
				Msg("pre-LIB entry")
		}
	}
}

// init restores the last LIB status by using the informations loaded from the
// DB.
func (s *Status) init() {
	if s.initialized {
		return
	}

	s.bestBlock = bootState.bestBlock()

	genesisBlock := bootState.genesisBlock()
	s.pls.genesisInfo = &blockInfo{
		BlockHash: genesisBlock.ID(),
		BlockNo:   genesisBlock.BlockNo(),
	}

	s.pls.addConfirmInfo(s.bestBlock)

	s.lib = bootState.lib

	if len(bootState.plib) != 0 {
		s.pls.plib = bootState.plib
	}

	s.initialized = true
}

// Update updates the last irreversible block (LIB).
func (s *Status) Update(block *types.Block) {
	s.Lock()
	defer s.Unlock()

	s.init()

	curBestID := s.bestBlock.ID()
	if curBestID == block.PrevID() {
		s.pls.addConfirmInfo(block)

		logger.Debug().
			Str("block hash", block.ID()).
			Uint64("block no", block.BlockNo()).
			Msg("update LIB status")

		// Block connected
		if lib := s.pls.update(); lib != nil {
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

func (bs *bootingStatus) bestBlock() *types.Block {
	return bs.best
}

func (bs *bootingStatus) genesisBlock() *types.Block {
	return bs.genesis
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

// Save saves the consensus status information for the later recovery.
func (s *Status) Save(tx db.Transaction) error {
	if len(s.pls.plib) != 0 {
		buf, err := encode(s.pls.plib)
		if err != nil {
			return err
		}
		plib := buf.Bytes()

		tx.Set(statusKeyPreLIB, plib)
	}

	if s.lib != nil {
		buf, err := encode(s.lib)
		if err != nil {
			return err
		}
		lib := buf.Bytes()

		tx.Set(statusKeyLIB, lib)
	}

	return nil
}

func encode(e interface{}) (bytes.Buffer, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(e)

	return buf, err
}

func decode(r io.Reader, e interface{}) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(e)
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
