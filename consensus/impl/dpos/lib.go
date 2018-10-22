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

type bpPlm map[string]*plInfo

func (plm bpPlm) set(bpID string, pl *plInfo) {
	plm[bpID] = pl
	logger.Debug().Str("BP", bpID).
		Str("hash", pl.Plib.BlockHash).Uint64("no", pl.Plib.BlockNo).
		Str("hash", pl.PlibBy.BlockHash).Uint64("no", pl.PlibBy.BlockNo).
		Msg("proposed LIB map updated")
}

type pLibStatus struct {
	genesisInfo      *blockInfo
	confirmsRequired uint16
	confirms         *list.List
	plm              bpPlm // BP-wise proposed LIB map
}

func newPlibStatus(confirmsRequired uint16) *pLibStatus {
	pls := &pLibStatus{
		confirmsRequired: confirmsRequired,
		confirms:         list.New(),
		plm:              make(bpPlm),
	}

	return pls
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
		pls.updatePreLIB(ci.BPID,
			&plInfo{
				Plib:   pls.genesisInfo,
				PlibBy: pls.genesisInfo,
			},
		)
	}

	logger.Debug().Str("BP", ci.BPID).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Msg("new confirm info added")
}

func (pls *pLibStatus) update() *blockInfo {
	if bpID, pl := pls.getPreLIB(); pl != nil {
		pls.updatePreLIB(bpID, pl)

		return pls.calcLIB()
	}
	return nil
}

func (pls *pLibStatus) updatePreLIB(bpID string, pl *plInfo) {
	pls.plm.set(bpID, pl)
}

func (pls *pLibStatus) getPreLIB() (bpID string, pl *plInfo) {
	var (
		confirmed *blockInfo
		prev      *list.Element
		e         = pls.confirms.Back()
		cr        = cInfo(e).ConfirmRange
	)
	bpID = cInfo(e).BPID
	confirmedBy := cInfo(e).blockInfo

	for e != nil && cr > 0 {
		prev = e.Prev()
		cr--

		c := cInfo(e)
		c.confirmsLeft--
		if c.confirmsLeft == 0 {
			// proposed LIB info to return
			confirmed = c.bInfo()
			break
		}

		e = prev
	}

	if confirmed != nil {
		pl = &plInfo{Plib: confirmed, PlibBy: confirmedBy}
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

	// Rebuild confirms info & pre-LIB map from LIB + 1 and block based on
	// the blocks.
	if tmp := loadPlibStatus(lib, block); tmp != nil {
		if tmp.confirms.Len() > 0 {
			pls.confirms = tmp.confirms
		}
		for bpID, v := range tmp.plm {
			if v != nil && v.Plib.BlockNo > 0 {
				pls.plm[bpID] = v
			}
		}
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

func (pls *pLibStatus) gc(lib *blockInfo) {
	removeIf(pls.confirms,
		func(e *list.Element) bool {
			return cInfo(e).BlockNo <= lib.BlockNo
		})
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

	libInfos := make([]*plInfo, 0, len(pls.plm))
	for _, l := range pls.plm {
		if l != nil {
			libInfos = append(libInfos, l)
		}
	}

	if len(libInfos) == 0 {
		return nil
	}

	sort.Slice(libInfos, func(i, j int) bool {
		return libInfos[i].Plib.BlockNo < libInfos[j].Plib.BlockNo
	})

	// TODO: check the correctness of the formula.
	lib := libInfos[(len(libInfos)-1)/3]

	return lib.Plib
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

func (bi *blockInfo) Hash() string {
	if bi == nil {
		return "(nil)"
	}
	return bi.BlockHash
}

type plInfo struct {
	PlibBy *blockInfo // the block info by which a block becomes pre-LIB.
	Plib   *blockInfo // pre-LIB
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

type bootLoader struct {
	pls      *pLibStatus
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

	if plm := bs.loadPLM(); plm != nil {
		bs.pls.plm = plm
		logger.Debug().Int("len", len(plm)).Msg("pre-LIB loaded from DB")
		for id, p := range plm {
			if p == nil {
				continue
			}
			logger.Debug().Str("BPID", id).
				Str("confirmed hash", p.Plib.Hash()).
				Str("confirmedBy hash", p.PlibBy.Hash()).
				Msg("pre-LIB entry")
		}

	}

	bs.pls.load(bs.lib, bs.best)
}

func (bs *bootLoader) loadLIB(bi *blockInfo) error {
	return bs.decodeStatus(statusKeyLIB, bi)
}

func (bs *bootLoader) loadPLM() bpPlm {
	plm := make(bpPlm)
	if err := bs.decodeStatus(statusKeyPreLIB, &plm); err != nil {
		return nil
	}

	return plm
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

func loadPlibStatus(lib *blockInfo, blockEnd *types.Block) *pLibStatus {
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

	beginBlockNo := func() uint64 {
		// For the case where no pre-LIB map are correctly restored at a boot
		// time.
		if beg := lib.BlockNo - 2*consensusBlockCount(); beg > 0 {
			return beg
		}
		return 1
	}

	for i := beginBlockNo(); i <= end; i++ {
		block, err := libLoader.getBlock(i)
		if err != nil {
			// XXX Better error handling?!
			logger.Error().Err(err).Msg("failed to read block")
			return nil
		}
		pls.addConfirmInfo(block)
		pls.update()
	}

	return pls
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
