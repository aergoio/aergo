package dpos

import (
	"bytes"
	"container/list"
	"fmt"
	"sort"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
)

var (
	statusKeyLIB    = []byte("dposStatus.LIB")
	statusKeyPreLIB = []byte("dposStatus.PreLIB")
	libLoader       *bootLoader
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

type errInvalidLIB struct {
	lastHash string
	lastNo   uint64
	libHash  string
	libNo    uint64
}

func (e errInvalidLIB) Error() string {
	return fmt.Sprintf("The LIB (%v, %v) is inconsistent with the best block (%v, %v)",
		e.libNo, e.libHash, e.lastNo, e.lastHash)
}

type bpPlm map[string][]*blockInfo

func (plm bpPlm) trunc(bpID string, beg, end int) {
	var (
		entry = plm[bpID]
		max   = len(entry)
	)

	if beg < 0 {
		beg = 0
	}
	if end > max {
		end = max
	}
	if beg == 0 && end == max {
		// nothing to trucate
		return
	}

	plm[bpID] = entry[beg:end]
	logger.Debug().
		Str("BPID", bpID).Int("old len", max).Int("new len", len(plm[bpID])).
		Msg("PLM truncated")
}

func (plm bpPlm) add(bpID string, bi *blockInfo) {
	v := plm[bpID]
	old := len(v)
	plm[bpID] = append(v, bi)
	logger.Debug().Str("BP", bpID).
		Int("old len", old).Int("new len", len(plm[bpID])).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Msg("proposed LIB map updated")
}

type pLibStatus struct {
	genesisInfo      *blockInfo
	confirmsRequired uint16
	confirms         *list.List
	plm              bpPlm // BP-wise proposed LIB map
}

func newPlibStatus(confirmsRequired uint16) *pLibStatus {
	return &pLibStatus{
		confirmsRequired: confirmsRequired,
		confirms:         list.New(),
		plm:              make(bpPlm),
	}
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
	if _, exist := pls.plm[ci.BPID]; !exist {
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
	pls.plm[bpID] = append(pls.plm[bpID], bi)
	logger.Debug().Str("BP", bpID).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Msg("proposed LIB map updated")
}

func (pls *pLibStatus) getPreLIB() (bpID string, bi *blockInfo) {
	var (
		prev *list.Element
		e    = pls.confirms.Back()
		cr   = cInfo(e).ConfirmRange
	)
	bpID = cInfo(e).BPID

	for e != nil && cr > 0 {
		prev = e.Prev()
		cr--

		c := cInfo(e)
		c.confirmsLeft--
		if c.confirmsLeft == 0 {
			// proposed LIB info to return
			bi = c.bInfo()
			break
		}

		e = prev
	}

	return
}

func (pls *pLibStatus) rollbackStatusTo(block *types.Block, lib *blockInfo) error {
	var (
		targetBlockNo = block.BlockNo()
	)

	logger.Debug().
		Uint64("target no", targetBlockNo).Int("confirms len", pls.confirms.Len()).
		Msg("start LIB status rollback")

	pls.load(lib, block)

	return nil
}

func (pls *pLibStatus) load(lib *blockInfo, block *types.Block) {
	// Remove all the previous confirmation info.
	if pls.confirms.Len() > 0 {
		pls.confirms.Init()
	}

	// Rebuild confirms info from LIB + 1 and block.
	if confirms := loadConfirms(lib, block); confirms != nil {
		pls.confirms = confirms

		// Rollback the pre-LIB map based on the new confirms list. -- During
		// rollback, no new pre-LIBs are created. Only some of the existing pre-LIB
		// map entries may be rollback to the previous one.
		pls.rollbackPreLIBs()
	}
}

func (pls *pLibStatus) save(tx db.Transaction) error {
	if len(pls.plm) != 0 {
		buf, err := encode(pls.plm)
		if err != nil {
			return err
		}
		plib := buf.Bytes()

		tx.Set(statusKeyPreLIB, plib)
	}
	return nil
}

func (pls *pLibStatus) rollbackPreLIBs() {
	forEach(pls.confirms,
		func(e *list.Element) {
			pls.rollbackPreLIB(cInfo(e))
		},
	)
}

func (pls *pLibStatus) rollbackPreLIB(c *confirmInfo) {
	if pLib, exist := pls.plm[c.BPID]; exist {
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
		pls.plm.trunc(c.BPID, 0, purgeBeg)
	}
}

func (pls *pLibStatus) gc(lib *blockInfo) {
	removeIf(pls.confirms,
		func(e *list.Element) bool {
			return cInfo(e).BlockNo <= lib.BlockNo
		})

	for bpID, pl := range pls.plm {
		var beg int
		for i, l := range pl {
			if l.BlockNo > lib.BlockNo {
				beg = i
				break
			}
		}
		pls.plm.trunc(bpID, beg, len(pl))
	}
}

func removeIf(l *list.List, p func(e *list.Element) bool) {
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

func (c *confirmInfo) bInfo() *blockInfo {
	return c.blockInfo
}

func cInfo(e *list.Element) *confirmInfo {
	return e.Value.(*confirmInfo)
}

func (pls *pLibStatus) calcLIB() *blockInfo {
	if len(pls.plm) == 0 {
		return nil
	}

	libInfos := make([]*blockInfo, 0, len(pls.plm))
	for _, l := range pls.plm {
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

func (bi *blockInfo) save(tx db.Transaction) error {
	if bi != nil {
		buf, err := encode(bi)
		if err != nil {
			return err
		}
		bi := buf.Bytes()

		tx.Set(statusKeyLIB, bi)
	}
	return nil
}

type bootLoader struct {
	plib     bpPlm
	lib      *blockInfo
	best     *types.Block
	genesis  *types.Block
	confirms *list.List

	get      func([]byte) []byte
	getBlock func(types.BlockNo) (*types.Block, error)
}

func (bs *bootLoader) load() {
	if err := bs.loadLIB(bs.lib); err == nil {
		logger.Debug().Uint64("block no", bs.lib.BlockNo).
			Str("block hash", bs.lib.BlockHash).Msg("LIB loaded from DB")
	}

	if err := bs.loadPLIB(&bs.plib); err == nil {
		logger.Debug().Int("len", len(bs.plib)).Msg("pre-LIB loaded from DB")
		for id, p := range bs.plib {
			if len(p) == 0 {
				continue
			}
			logger.Debug().
				Str("BPID", id).Str("block hash", p[len(p)-1].BlockHash).
				Msg("pre-LIB entry")
		}
	}

	libLoader.loadConfirms()
}

func (bs *bootLoader) loadLIB(bi *blockInfo) error {
	return bs.decodeStatus(statusKeyLIB, bi)
}

func (bs *bootLoader) loadPLIB(plib *bpPlm) error {
	return bs.decodeStatus(statusKeyPreLIB, plib)
}

func (bs *bootLoader) decodeStatus(key []byte, dst interface{}) error {
	value := bs.get(key)
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

func (bs *bootLoader) loadConfirms() {
	if confirms := loadConfirms(bs.lib, bs.best); confirms != nil {
		bs.confirms = confirms
	}
}

func loadConfirms(lib *blockInfo, blockEnd *types.Block) *list.List {
	end := blockEnd.BlockNo()
	if lib.BlockNo == end {
		return nil
	} else if lib.BlockNo > end {
		panic(errInvalidLIB{
			lastHash: blockEnd.ID(),
			lastNo:   end,
			libHash:  lib.BlockHash,
			libNo:    lib.BlockNo,
		})
	}

	pls := newPlibStatus(defaultConsensusCount)
	pls.genesisInfo = newBlockInfo(libLoader.genesis)

	beg := lib.BlockNo + 1
	for i := beg; i <= end; i++ {
		block, err := libLoader.getBlock(i)
		if err != nil {
			panic(err)
		}
		pls.addConfirmInfo(block)
		pls.update()
	}

	if pls.confirms.Len() > 0 {
		return pls.confirms
	}

	return nil
}

func (bs *bootLoader) bestBlock() *types.Block {
	return bs.best
}

func (bs *bootLoader) genesisBlock() *types.Block {
	return bs.genesis
}

func dumpConfirmInfo(name string, l *list.List) {
	forEach(l,
		func(e *list.Element) {
			logger.Debug().Str("confirm info", spew.Sdump(cInfo(e))).Msg(name)
		},
	)
}
