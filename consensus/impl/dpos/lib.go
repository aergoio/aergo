package dpos

import (
	"container/list"
	"fmt"
	"sort"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/types"
	"github.com/davecgh/go-spew/spew"
)

// LibStatusKey is the key when a LIB information is put into the chain DB.
var LibStatusKey = []byte("dpos.LibStatus")

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

type proposed map[string]*plInfo

func (pm proposed) set(bpID string, pl *plInfo) {
	pm[bpID] = pl
	logger.Debug().Str("BP", bpID).
		Str("hash", pl.Plib.BlockHash).Uint64("no", pl.Plib.BlockNo).
		Str("hash", pl.PlibBy.BlockHash).Uint64("no", pl.PlibBy.BlockNo).
		Msg("proposed LIB map updated")
}

func (pm proposed) gc(bps []string) {
	if len(bps) == 0 {
		return
	}
	// len(bps) must be larger than 0.
	filter := make(map[string]struct{})
	for _, bp := range bps {
		filter[bp] = struct{}{}
	}
	for bp := range pm {
		if _, ok := filter[bp]; !ok {
			delete(pm, bp)
		}
	}
}

type libStatus struct {
	Prpsd            proposed // BP-wise proposed LIB map
	Lib              *blockInfo
	LpbNo            types.BlockNo
	confirms         *list.List
	genesisInfo      *blockInfo
	bpid             string
	confirmsRequired uint16
}

func newLibStatus(bpCount uint16) *libStatus {
	ls := &libStatus{
		Prpsd:    make(proposed),
		Lib:      &blockInfo{},
		confirms: list.New(),
		bpid:     p2pkey.NodeSID(),
	}
	ls.setConfirmsRequired(bpCount)
	return ls
}

func (ls libStatus) lpbNo() types.BlockNo {
	return ls.LpbNo
}

func (ls libStatus) libNo() types.BlockNo {
	return ls.Lib.BlockNo
}

func (ls libStatus) lib() *blockInfo {
	return ls.Lib
}

func (ls *libStatus) addConfirmInfo(block *types.Block) {
	// Genesis block must not be added.
	if block.BlockNo() == 0 {
		return
	}

	ci := newConfirmInfo(block, ls.confirmsRequired)

	bi := ci.blockInfo

	if e := ls.confirms.PushBack(ci); e.Prev() != nil {
		prevBi := cInfo(e.Prev()).blockInfo
		if bi.BlockNo != prevBi.BlockNo+1 {
			logger.Error().
				Uint64("prev no", prevBi.BlockNo).Uint64("current no", bi.BlockNo).
				Msg("inconsistent confirm info found")
		}

	}

	// Initialize an empty pre-LIB map entry with genesis block info.
	if _, exist := ls.Prpsd[ci.bpid]; !exist {
		ls.updatePreLIB(ci.bpid,
			&plInfo{
				Plib:   ls.genesisInfo,
				PlibBy: ls.genesisInfo,
			},
		)
	}

	if ci.bpid == ls.bpid {
		ls.LpbNo = block.BlockNo()
	}

	logger.Debug().Str("BP", ci.bpid).
		Str("hash", bi.BlockHash).Uint64("no", bi.BlockNo).
		Uint64("range", ci.ConfirmRange).Uint16("confirms left", ci.confirmsLeft).
		Msg("new confirm info added")
}

func (ls *libStatus) update() *blockInfo {
	if bpID, pl := ls.getPreLIB(); pl != nil {
		ls.updatePreLIB(bpID, pl)

		return ls.calcLIB()
	}
	return nil
}

func (ls *libStatus) updatePreLIB(bpID string, pl *plInfo) {
	ls.Prpsd.set(bpID, pl)
}

func (ls *libStatus) getPreLIB() (bpID string, pl *plInfo) {
	var (
		confirmed   *blockInfo
		last        = ls.confirms.Back()
		lastCi      = cInfo(last)
		confirmedBy = lastCi.blockInfo
	)

	min := lastCi.BlockNo - lastCi.ConfirmRange + 1
	max := lastCi.BlockNo
	for e := last; e != nil; e = e.Prev() {
		c := cInfo(e)
		if c.BlockNo >= min && c.BlockNo <= max {
			c.confirmsLeft--
		}

		if c.confirmsLeft == 0 {
			confirmed = c.bInfo()
			break
		}
	}

	if confirmed != nil {
		bpID = lastCi.bpid
		pl = &plInfo{Plib: confirmed, PlibBy: confirmedBy}
	}

	return
}

func (ls *libStatus) begRecoBlockNo(endBlockNo types.BlockNo) types.BlockNo {
	offset := 3 * types.BlockNo(ls.confirmsRequired)

	libBlockNo := ls.Lib.BlockNo

	// To reduce IO operation
	begNo := endBlockNo
	if begNo < libBlockNo {
		begNo = libBlockNo
	}

	if begNo > offset {
		begNo -= offset
	} else {
		begNo = 1
	}

	return begNo
}

func (ls *libStatus) rollbackStatusTo(block *types.Block, lib *blockInfo) error {
	targetBlockNo := block.BlockNo()

	logger.Debug().
		Uint64("target no", targetBlockNo).Int("confirms len", ls.confirms.Len()).
		Msg("start LIB status rollback")

	ls.load(targetBlockNo)

	return nil
}

func (ls *libStatus) setConfirmsRequired(bpCount uint16) {
	consensusBlockCount := func(bpCount uint16) uint16 {
		return bpCount*2/3 + 1
	}
	ls.confirmsRequired = consensusBlockCount(bpCount)
}

func (ls *libStatus) load(endBlockNo types.BlockNo) {
	// Remove all the previous confirmation info.
	if ls.confirms.Len() > 0 {
		ls.confirms.Init()
	}

	// Nothing left for the genesis block.
	if endBlockNo == 0 {
		return
	}

	begBlockNo := ls.begRecoBlockNo(endBlockNo)

	// Rebuild confirms info & pre-LIB map from LIB + 1 and block based on
	// the blocks.
	if tmp := loadPlibStatus(begBlockNo, endBlockNo, ls.confirmsRequired); tmp != nil {
		if tmp.confirms.Len() > 0 {
			ls.confirms = tmp.confirms
		}
		for bpID, v := range tmp.Prpsd {
			if v != nil && v.Plib.BlockNo > 0 {
				ls.Prpsd[bpID] = v
			}
		}
	}
}

func (ls *libStatus) save(tx consensus.TxWriter) error {
	b, err := common.GobEncode(ls)
	if err != nil {
		return err
	}

	tx.Set(LibStatusKey, b)

	logger.Debug().Int("proposed lib len", len(ls.Prpsd)).Msg("lib status stored to DB")

	return nil
}

func reset(tx db.Transaction) {
	tx.Delete(LibStatusKey)
}

func (ls *libStatus) gc(bps []string) {
	// GC based on the LIB no
	if ls.Lib != nil {
		removeIf(ls.confirms,
			func(e *list.Element) bool {
				return cInfo(e).BlockNo <= ls.Lib.BlockNo
			},
			func(e *list.Element) bool {
				return cInfo(e).BlockNo > ls.Lib.BlockNo
			},
		)
	}
	// GC based on the element no
	limitConfirms := ls.gcNumLimit()
	nRemoved := 0
	for ls.confirms.Len() > limitConfirms {
		ls.confirms.Remove(ls.confirms.Front())
		nRemoved++
	}

	if nRemoved > 0 {
		logger.Debug().Int("len", ls.confirms.Len()).
			Int("limit", limitConfirms).Int("removed", nRemoved).
			Msg("number-based GC done for confirms list")
	}

	ls.Prpsd.gc(bps)
}

func (ls libStatus) gcNumLimit() int {
	return int(ls.confirmsRequired * 3)
}

func removeIf(l *list.List, p func(e *list.Element) bool, bc func(e *list.Element) bool) {
	forEachCond(l,
		func(e *list.Element) {
			if p(e) {
				l.Remove(e)
			}
		},
		bc,
	)
}

func forEachCond(l *list.List, f func(e *list.Element), bc func(e *list.Element) bool) {
	e := l.Front()
	for e != nil {
		next := e.Next()
		if bc(e) {
			break
		}
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

func (ls *libStatus) calcLIB() *blockInfo {
	if len(ls.Prpsd) == 0 {
		return nil
	}

	libInfos := make([]*plInfo, 0, len(ls.Prpsd))
	for _, l := range ls.Prpsd {
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
	bpid         string
	confirmsLeft uint16
}

func newConfirmInfo(block *types.Block, confirmsRequired uint16) *confirmInfo {
	return &confirmInfo{
		bpid:         block.BPID2Str(),
		blockInfo:    newBlockInfo(block),
		confirmsLeft: confirmsRequired,
	}
}

func (bs *bootLoader) loadLibStatus() *libStatus {
	pls := newLibStatus(bs.confirmsRequired)
	if err := bs.decodeStatus(LibStatusKey, pls); err != nil {
		return nil
	}
	pls.load(bs.best.BlockNo())

	return pls
}

func (bs *bootLoader) decodeStatus(key []byte, dst interface{}) error {
	value := bs.cdb.Get(key)
	if len(value) == 0 {
		return fmt.Errorf("LIB status not found: key = %v", string(key))
	}

	err := common.GobDecode(value, dst)
	if err != nil {
		logger.Panic().Err(err).Str("key", string(key)).
			Msg("failed to decode DPoS status")
	}
	return nil
}

func loadPlibStatus(begBlockNo, endBlockNo types.BlockNo, confirmsRequired uint16) *libStatus {
	if begBlockNo == endBlockNo {
		return nil
	} else if begBlockNo > endBlockNo {
		logger.Info().Uint64("beg", begBlockNo).Uint64("end", endBlockNo).
			Msg("skip pre-LIB status recovery due to the invalid block range")
		return nil
	} else if begBlockNo == 0 {
		begBlockNo = 1
	}

	pls := newLibStatus(confirmsRequired)
	pls.genesisInfo = newBlockInfo(bsLoader.genesis)

	logger.Debug().Uint64("beginning", begBlockNo).Uint64("ending", endBlockNo).
		Msg("restore pre-LIB status from blocks")
	for i := begBlockNo; i <= endBlockNo; i++ {
		block, err := bsLoader.cdb.GetBlockByNo(i)
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

func (bs *bootLoader) lpbNo() types.BlockNo {
	return bs.ls.lpbNo()
}

func dumpConfirmInfo(name string, l *list.List) {
	forEach(l,
		func(e *list.Element) {
			logger.Debug().Str("confirm info", spew.Sdump(cInfo(e))).Msg(name)
		},
	)
}

func forEach(l *list.List, f func(e *list.Element)) {
	for e := l.Front(); e != nil; e = e.Next() {
		f(e)
	}
}
