/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package bp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/davecgh/go-spew/spew"
)

const (
	max = 100

	// New BPs are elected every maxBpLimit blocks.
	electionPeriod = types.BlockNo(max)
)

var (
	logger  = log.NewLogger("bp")
	errNoBP = errors.New("no block producers found in the block chain")

	genesisBpList []string
)

type errBpSize struct {
	required uint16
	given    uint16
}

func (e errBpSize) Error() string {
	return fmt.Sprintf("insufficient or redundant block producers  - %v (required - %v)", e.given, e.required)
}

// Max returns the maximum number of active block producers.
func Max() uint16 {
	return max
}

// ClusterMember is an interface which corresponds to BP member update.
type ClusterMember interface {
	Size() uint16
	Update(ids []string) error
}

// Cluster represents a cluster of block producers.
type Cluster struct {
	sync.RWMutex
	size   uint16
	member map[Index]*blockProducer
	index  map[types.PeerID]Index

	cdb consensus.ChainDB
}

// blockProducer represents one member in the block producer cluster.
type blockProducer struct {
	id types.PeerID
}

// NewCluster returns a new bp.Cluster.
func NewCluster(cdb consensus.ChainDB) (*Cluster, error) {
	c := &Cluster{cdb: cdb}

	if err := c.init(); err != nil {
		return nil, err
	}

	return c, nil
}

func newBlockProducer(id types.PeerID) *blockProducer {
	return &blockProducer{id: id}
}

func (c *Cluster) init() error {
	if c.cdb == nil {
		return errNoBP
	}

	if genesisBpList = c.genesisBpList(); len(genesisBpList) == 0 {
		return errNoBP
	}

	// The total BP count is determined by the genesis info and afterwards it
	// remains the same.
	c.size = uint16(len(genesisBpList))

	// The boot time BP member loading is later performed along with DPoS
	// status initialization.

	return nil
}

func bootstrapHeight() types.BlockNo {
	return getElectionPeriod() * 3
}

func (c *Cluster) genesisBpList() []string {
	genesis := c.cdb.GetGenesisInfo()
	if genesis != nil {
		logger.Debug().Str("genesis", spew.Sdump(genesis)).Msg("genesis info loaded")
		// Prefer BPs from the GenesisInfo. Overwrite.
		if len(genesis.BPs) > 0 {
			logger.Debug().Msg("use BPs from the genesis info")
			for i, bp := range genesis.BPs {
				logger.Debug().Int("no", i).Str("ID", bp).Msg("Genesis BP")
			}
			return genesis.BPs
		}
	}
	return nil
}

// BPs returns BP information about each BP in JSON.
func (c *Cluster) BPs() []string {
	c.RLock()
	defer c.RUnlock()

	if c == nil || c.getSize() == 0 || len(c.member) != int(c.getSize()) {
		return nil
	}
	bps := make([]string, c.getSize())
	for i, bp := range c.member {
		p := &struct {
			Index  string
			PeerID string
		}{
			Index:  strconv.FormatUint(uint64(i), 10),
			PeerID: bp.id.Pretty(),
		}

		m, err := json.Marshal(p)
		if err != nil {
			bps = nil
			break
		}
		bps[int(i)] = string(m)
	}
	return bps
}

// Update updates old cluster index by using ids.
func (c *Cluster) Update(ids []string) error {
	c.Lock()
	defer c.Unlock()

	bpMember := make(map[Index]*blockProducer)
	bpIndex := make(map[types.PeerID]Index)

	for i, id := range ids {
		bpID, err := types.IDB58Decode(id)
		if err != nil {
			return fmt.Errorf("invalid node ID[%d]: %s", i, err.Error())
		}

		var index Index
		if index, err = newIndex(i); err != nil {
			return err
		}

		bpMember[index] = newBlockProducer(bpID)
		bpIndex[bpID] = index
	}

	c.size = uint16(len(bpMember))
	c.member = bpMember
	c.index = bpIndex

	logger.Debug().Msgf("BP list updated. member: %v", ids)

	return nil
}

// Size returns c.size.
func (c *Cluster) Size() uint16 {
	c.RLock()
	defer c.RUnlock()
	return c.getSize()
}

func (c *Cluster) getSize() uint16 {
	return c.size
}

// Index is a type for a block producer index.
type Index uint16

// indexNil is the nil value for BpIndex type
const (
	indexNil = Index(math.MaxUint16)
	indexMax = indexNil - 1
)

func newIndex(i int) (Index, error) {
	if i > int(indexMax) {
		return indexNil, fmt.Errorf("BP index [%v] is too big", i)
	}
	return Index(i), nil
}

// IsNil reports whether idx is nil or not.
func (idx Index) IsNil() bool {
	return idx == indexNil
}

// BpIndex2ID returns the ID corresponding to idx.
func (c *Cluster) BpIndex2ID(bpIdx Index) (types.PeerID, bool) {
	c.Lock()
	defer c.Unlock()

	if bp, exist := c.member[bpIdx]; exist {
		return bp.id, exist
	}
	return types.PeerID(""), false
}

// BpID2Index returns the index corresponding to id.
func (c *Cluster) BpID2Index(id types.PeerID) Index {
	c.Lock()
	defer c.Unlock()
	idx, exist := c.index[id]
	if exist {
		return idx
	}

	return indexNil
}

// Has reports whether c includes id or not
func (c *Cluster) Has(id types.PeerID) bool {
	c.Lock()
	defer c.Unlock()
	_, exist := c.index[id]
	return exist
}

// Snapshot represents the set of the elected BP at refBlockNo.
type Snapshot struct {
	RefBlockNo types.BlockNo
	List       []string
}

// NewSnapshot returns a Snapshot corresponding to blockNo and period.
func NewSnapshot(blockNo types.BlockNo, bps []string) (*Snapshot, error) {
	if !isSnapPeriod(blockNo) {
		return nil, fmt.Errorf("%v is not inconsistent with period %v", blockNo, getElectionPeriod())
	}
	return &Snapshot{RefBlockNo: blockNo, List: bps}, nil
}

func snapBlockNo(blockNo types.BlockNo) types.BlockNo {
	if blockNo < bootstrapHeight() {
		return 0
	}
	return (blockNo/getElectionPeriod() - 1) * getElectionPeriod()
}

func isSnapPeriod(blockNo types.BlockNo) bool {
	// The current snapshot period is the total BP count.
	return blockNo%getElectionPeriod() == 0
}

// Key returns the properly prefixed key corresponding to s.
func (s *Snapshot) Key() []byte {
	return buildKey(s.RefBlockNo)
}

func buildKey(blockNo types.BlockNo) []byte {
	const bpListPrefix = "dpos.BpList"

	return []byte(fmt.Sprintf("%v.%v", bpListPrefix, blockNo))
}

// Value returns s.list.
func (s *Snapshot) Value() []byte {
	b, err := common.GobEncode(s.List)
	if err != nil {
		logger.Debug().Err(err).Msg("BP list encoding failed")
		return nil
	}
	return b
}

const (
	opNil = iota
	opAdd
	opDel
)

type journal struct {
	op      int
	blockNo types.BlockNo
}

// Snapshots is a map from block no to *Snapshot.
type Snapshots struct {
	snaps         map[types.BlockNo]*Snapshot
	maxRefBlockNo types.BlockNo
	cm            ClusterMember
	cdb           consensus.ChainDB
	sdb           *state.ChainStateDB
}

// NewSnapshots returns a new Snapshots.
func NewSnapshots(c ClusterMember, cdb consensus.ChainDB, sdb *state.ChainStateDB) *Snapshots {
	snap := &Snapshots{
		snaps: make(map[types.BlockNo]*Snapshot),
		cm:    c,
		cdb:   cdb,
		sdb:   sdb,
	}

	// To avoid a unit test failure.
	if cdb == nil {
		return snap
	}

	// Initialize the BP cluster members.
	if block, err := cdb.GetBestBlock(); err == nil {
		snap.UpdateCluster(block.BlockNo())
	} else {
		logger.Panic().Err(err).Msg("Failed to get the best block")
	}

	return snap
}

// NeedToRefresh reports whether blockNo corresponds to a BP regime change
// point.
func (sn *Snapshots) NeedToRefresh(blockNo types.BlockNo) bool {
	return blockNo%getElectionPeriod() == 0
}

// AddSnapshot add a new BP list corresponding to refBlockNO to sn.
func (sn *Snapshots) AddSnapshot(refBlockNo types.BlockNo) ([]string, error) {
	// Reorganization!!!
	if sn.maxRefBlockNo > refBlockNo {
		sn.reset()
	}

	// Add BP list every 'sn.bpCount()'th block.
	if sn.sdb == nil || !isSnapPeriod(refBlockNo) || refBlockNo == 0 {
		return nil, nil
	}

	var (
		bps []string
		err error
	)

	if bps, err = sn.gatherRankers(); err != nil {
		return nil, err
	}

	if err := sn.add(refBlockNo, bps); err != nil {
		return nil, err
	}

	if sn.NeedToRefresh(refBlockNo) {
		// Return bps from UpdateCluster if the cluster updated.
		bps = sn.UpdateCluster(refBlockNo)
	}

	sn.gc(refBlockNo)

	return bps, nil
}

func (sn *Snapshots) gatherRankers() ([]string, error) {
	return system.GetRankers(sn.sdb)
}

// UpdateCluster updates the current BP list by the ones corresponding to
// blockNo.
func (sn *Snapshots) UpdateCluster(blockNo types.BlockNo) (s []string) {
	var err error

	if s, err = sn.getCurrentCluster(blockNo); err == nil {
		logger.Debug().Uint64("cur block no", blockNo).Msg("get BP list snapshot")
		err = sn.cm.Update(s)
	}

	if err != nil {
		logger.Debug().Err(err).Msg("skip BP member update")
		return nil
	}

	return
}

func (sn *Snapshots) Size() uint16 {
	return sn.cm.Size()
}

func (sn *Snapshots) reset() {
	sn.snaps = make(map[types.BlockNo]*Snapshot)
}

// add adds a new BP snapshot to snap.
func (sn *Snapshots) add(refBlockNo types.BlockNo, bps []string) error {
	var (
		s   *Snapshot
		err error
	)

	if s, err = NewSnapshot(refBlockNo, bps); err != nil {
		return err
	}

	sn.snaps[refBlockNo] = s

	logger.Debug().Uint64("ref block no", refBlockNo).Msgf("BP snapshot added: %v", bps)

	return nil
}

// del removes a snapshot corresponding to refBlockNo from sn.snaps.
func (sn *Snapshots) del(refBlockNo types.BlockNo) error {
	if _, exist := sn.snaps[refBlockNo]; !exist {
		logger.Debug().Uint64("ref block no", refBlockNo).Msg("no such an entry in BP snapshots. ignored.")
		return nil
	}

	delete(sn.snaps, refBlockNo)

	logger.Debug().Uint64("block no", refBlockNo).Int("len", len(sn.snaps)).Msg("BP snapshot removed")

	return nil
}

// gc remove all the snapshots less than blockNo
func (sn *Snapshots) gc(blockNo types.BlockNo) {
	gcPeriod := sn.gcPeriod()

	var gcBlockNo types.BlockNo
	if blockNo > gcPeriod {
		gcBlockNo = blockNo - gcPeriod
	}

	for h := range sn.snaps {
		if h < gcBlockNo {
			sn.del(h)
		}
	}
}

func getElectionPeriod() types.BlockNo {
	return electionPeriod
}

func (sn Snapshots) period() types.BlockNo {
	return getElectionPeriod()
}

func (sn Snapshots) gcPeriod() types.BlockNo {
	return 2 * sn.period()
}

// getCurrentCluster returns the BP snapshot corresponding to blockNo.
func (sn *Snapshots) getCurrentCluster(blockNo types.BlockNo) ([]string, error) {
	refBlockNo := snapBlockNo(blockNo)
	if refBlockNo == 0 {
		return genesisBpList, nil
	}

	if s, exist := sn.snaps[refBlockNo]; exist {
		return s.List, nil
	}

	return sn.loadClusterSnapshot(blockNo)
}

func (sn *Snapshots) loadClusterSnapshot(blockNo types.BlockNo) ([]string, error) {
	var (
		block *types.Block
		err   error
	)

	block, err = sn.cdb.GetBlockByNo(snapBlockNo(blockNo))
	if err != nil {
		return nil, err
	}

	stateDB := sn.sdb.OpenNewStateDB(block.GetHeader().GetBlocksRootHash())

	return system.GetRankers(stateDB)
}
