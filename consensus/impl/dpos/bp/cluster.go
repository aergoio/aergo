/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package bp

import (
	"errors"
	"fmt"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-peer"
)

var (
	logger = log.NewLogger("bp")

	errNoBP = errors.New("no block producers found in the block chain")
)

type errBpSize struct {
	required uint16
	given    uint16
}

func (e errBpSize) Error() string {
	return fmt.Sprintf("insufficient or redundant block producers  - %v (required - %v)", e.given, e.required)
}

// Cluster represents a cluster of block producers.
type Cluster struct {
	size   uint16
	member map[uint16]*blockProducer
	index  map[peer.ID]uint16

	cdb consensus.ChainDbReader
}

// blockProducer represents one member in the block producer cluster.
type blockProducer struct {
	id peer.ID
}

// NewCluster returns a new bp.Cluster.
func NewCluster(cfg *config.ConsensusConfig, cdb consensus.ChainDbReader) (*Cluster, error) {
	c := &Cluster{cdb: cdb}

	if err := c.init(); err != nil {
		return nil, err
	}

	return c, nil
}

func newBlockProducer(id peer.ID) *blockProducer {
	return &blockProducer{id: id}
}

func (c *Cluster) init() error {
	if c.cdb == nil {
		return errNoBP
	}

	var bps0, bps []string

	if bps0 = c.genesisBpList(); len(bps0) == 0 {
		return errNoBP
	}

	// The total BP count is determined by the genesis info and afterwards it
	// remains the same.
	c.size = uint16(len(bps0))

	var (
		bestBlock *types.Block
		err       error
	)

	if bestBlock, err = c.cdb.GetBestBlock(); err != nil {
		return err
	}

	// During the initial boostrapping period, the BPs given by the genesis
	// info is used.
	if bestBlock.BlockNo() <= c.bootstrapHeight() {
		bps = bps0
	}

	bps = c.currentBpList()

	if err := c.Update(bps); err != nil {
		return err
	}

	return nil
}

func (c Cluster) bootstrapHeight() types.BlockNo {
	const period = 10

	return types.BlockNo(c.Size()) * period
}

func (c *Cluster) genesisBpList() []string {
	genesis := c.cdb.GetGenesisInfo()
	if genesis != nil {
		logger.Debug().Str("genesis", spew.Sdump(genesis)).Msg("genesis info loaded")
		// Prefer BPs from the GenesisInfo. Overwrite.
		if len(genesis.BPs) > 0 {
			logger.Debug().Msg("use BPs from the genesis info")
			for i, bp := range genesis.BPs {
				logger.Debug().Int("no", i).Str("ID", bp).Msg("BP")
			}
			return genesis.BPs
		}
	}
	return nil
}

func (c *Cluster) currentBpList() []string {
	// TODO: Get the elected BPs instead of the genesis BPs.
	return c.genesisBpList()
}

// Update updates old cluster index by using ids.
func (c *Cluster) Update(ids []string) error {
	c.member = make(map[uint16]*blockProducer)
	c.index = make(map[peer.ID]uint16)

	for i, id := range ids {
		bpID, err := peer.IDB58Decode(id)
		if err != nil {
			return fmt.Errorf("invalid node ID[%d]: %s", i, err.Error())
		}

		index := uint16(i)
		c.member[index] = newBlockProducer(bpID)
		c.index[bpID] = index
	}

	if len(c.index) != int(c.size) {
		return errBpSize{required: c.size, given: uint16(len(ids))}
	}

	return nil
}

// Size returns c.size.
func (c *Cluster) Size() uint16 {
	return c.size
}

// BpIndex2ID returns the ID correspinding to idx.
func (c *Cluster) BpIndex2ID(idx uint16) (peer.ID, bool) {
	if bp, exist := c.member[idx]; exist {
		return bp.id, exist
	}
	return peer.ID(""), false
}

// BpID2Index returns the index corresponding to id.
func (c *Cluster) BpID2Index(id peer.ID) (uint16, bool) {
	idx, exist := c.index[id]
	return idx, exist
}

// Has reports whether c includes id or not
func (c *Cluster) Has(id peer.ID) bool {
	_, exist := c.index[id]
	return exist
}

// Snapshot represents the set of the elected BP at refBlockNo.
type Snapshot struct {
	refBlockNo types.BlockNo
	list       []string
}

// NewSnapshot returns a Snapshot corresponding to blockNo and period.
func NewSnapshot(blockNo, period types.BlockNo, list []string) (*Snapshot, error) {
	if blockNo%period != 0 {
		return nil, fmt.Errorf("block no %v is inconsistent with period %v", blockNo, period)
	}
	return &Snapshot{refBlockNo: blockNo, list: list}, nil
}

// Key returns the properly prefixed key corresponding to s.
func (s *Snapshot) Key() []byte {
	return buildKey(s.refBlockNo)
}

func buildKey(blockNo types.BlockNo) []byte {
	const bpListPrefix = "dpos.BpList"

	return []byte(fmt.Sprintf("%v.%v", bpListPrefix, blockNo))
}

// Value returns s.list.
func (s *Snapshot) Value() []byte {
	b, err := common.GobEncode(s.list)
	if err != nil {
		return nil
	}
	return b
}
