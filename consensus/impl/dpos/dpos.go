/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/davecgh/go-spew/spew"
	peer "github.com/libp2p/go-libp2p-peer"
)

var (
	logger = log.NewLogger("dpos")

	// blockProducers is the number of block producers
	blockProducers        uint16
	defaultConsensusCount uint16

	lastJob = &lastSlot{}
)

type lastSlot struct {
	//	sync.Mutex
	s *slot.Slot
}

func (l *lastSlot) get() *slot.Slot {
	//	l.Lock()
	//	defer l.Unlock()
	return l.s
}

func (l *lastSlot) set(s *slot.Slot) {
	//	l.Lock()
	//	defer l.Unlock()
	l.s = s
}

// DPoS is the main data structure of DPoS consensus
type DPoS struct {
	*Status
	*component.ComponentHub
	bpc  *bp.Cluster
	bf   *BlockFactory
	quit chan interface{}
	ca   types.ChainAccessor
}

// Status shows DPoS consensus's current status
type bpInfo struct {
	bestBlock *types.Block
	slot      *slot.Slot
	ca        types.ChainAccessor
}

func (bi *bpInfo) updateBestBLock() *types.Block {
	block, _ := bi.ca.GetBestBlock()
	if block != nil {
		bi.bestBlock = block
	}

	return block
}

// New returns a new DPos object
func New(cfg *config.Config, cdb consensus.ChainDbReader, hub *component.ComponentHub) (consensus.Consensus, error) {
	genesis := cdb.GetGenesisInfo()
	if genesis != nil {
		logger.Debug().Str("genesis", spew.Sdump(genesis)).Msg("genesis info loaded")
		bpCount := len(genesis.BPs)
		// Prefer BPs from the GenesisInfo. Overwrite.
		if bpCount > 0 {
			logger.Debug().Msg("use BPs from the genesis info")
			for i, bp := range genesis.BPs {
				logger.Debug().Int("no", i).Str("ID", bp).Msg("BP")
			}
			cfg.Consensus.BpIds = genesis.BPs
			cfg.Consensus.DposBpNumber = uint16(bpCount)
		}
	}

	Init(cfg.Consensus)

	bpc, err := bp.NewCluster(cfg.Consensus.BpIds, blockProducers)
	if err != nil {
		return nil, err
	}

	quitC := make(chan interface{})

	return &DPoS{
		Status:       NewStatus(defaultConsensusCount, cdb),
		ComponentHub: hub,
		bpc:          bpc,
		bf:           NewBlockFactory(hub, quitC),
		quit:         quitC,
	}, nil
}

// Init initilizes the DPoS parameters.
func Init(cfg *config.ConsensusConfig) {
	consensus.InitBlockInterval(cfg.BlockInterval)

	blockProducers = cfg.DposBpNumber
	defaultConsensusCount = blockProducers*2/3 + 1
	slot.Init(cfg.BlockInterval, blockProducers)
}

func consensusBlockCount() uint64 {
	return uint64(defaultConsensusCount)
}

// Ticker returns a time.Ticker for the main consensus loop.
func (dpos *DPoS) Ticker() *time.Ticker {
	return time.NewTicker(tickDuration())
}

func tickDuration() time.Duration {
	return consensus.BlockInterval / 100
}

// QueueJob send a block triggering information to jq.
func (dpos *DPoS) QueueJob(now time.Time, jq chan<- interface{}) {
	bpi := dpos.getBpInfo(now)
	if bpi != nil {
		jq <- bpi
		lastJob.set(bpi.slot)
	}
}

// BlockFactory returns the BlockFactory interface in dpos.
func (dpos *DPoS) BlockFactory() consensus.BlockFactory {
	return dpos.bf
}

// SetChainAccessor sets dpost.ca to chainAccessor
func (dpos *DPoS) SetChainAccessor(chainAccessor types.ChainAccessor) {
	dpos.ca = chainAccessor
}

// SetStateDB sets sdb to the corresponding field of DPoS. This method is
// called only once during the boot sequence.
func (dpos *DPoS) SetStateDB(sdb *state.ChainStateDB) {
	dpos.bf.sdb = sdb
}

// IsTransactionValid checks the DPoS consensus level validity of a transaction
func (dpos *DPoS) IsTransactionValid(tx *types.Tx) bool {
	// TODO: put a transaction validity check code here.
	return true
}

// QuitChan returns the channel from which consensus-related goroutines check when
// shutdown is initiated.
func (dpos *DPoS) QuitChan() chan interface{} {
	return dpos.quit
}

func (dpos *DPoS) bpid() peer.ID {
	return p2p.NodeID()
}

// IsBlockValid checks the DPoS consensus level validity of a block
func (dpos *DPoS) IsBlockValid(block *types.Block, bestBlock *types.Block) error {
	id, err := block.BPID()
	if err != nil {
		return &consensus.ErrorConsensus{Msg: "bad public key in block", Err: err}
	}

	ns := block.GetHeader().GetTimestamp()
	idx, ok := dpos.bpc.BpID2Index(id)
	s := slot.NewFromUnixNano(ns)
	// Check whether the BP ID is one of the current BP members and its
	// corresponding BP index is consistent with the block timestamp.
	if !ok || !s.IsFor(idx) {
		return &consensus.ErrorConsensus{
			Msg: fmt.Sprintf("BP %v (idx: %v) is not permitted for the time slot %v (%v)",
				block.BPID2Str(), idx, time.Unix(0, ns), s.NextBpIndex()),
		}
	}

	valid, err := block.VerifySign()
	if !valid {
		return &consensus.ErrorConsensus{Msg: "bad block signature", Err: err}
	}

	return nil
}

func (dpos *DPoS) bpIdx() uint16 {
	idx, exist := dpos.bpc.BpID2Index(dpos.bpid())
	if !exist {
		logger.Fatal().Str("id", enc.ToString([]byte(dpos.bpid()))).Msg("BP has no correct BP membership")
	}

	return idx
}

func (dpos *DPoS) getBpInfo(now time.Time) *bpInfo {
	s := slot.Time(now)

	if !s.IsFor(dpos.bpIdx()) {
		return nil
	}

	// already queued slot.
	if slot.Equal(s, lastJob.get()) {
		return nil
	}

	block, _ := dpos.ca.GetBestBlock()
	logger.Debug().Str("best", block.ID()).Uint64("no", block.GetHeader().GetBlockNo()).
		Msg("GetBestBlock from BP")
	if block == nil {
		return nil
	}

	if !isBpTiming(block, s) {
		return nil
	}

	return &bpInfo{
		bestBlock: block,
		slot:      s,
		ca:        dpos.ca,
	}
}

func isBpTiming(block *types.Block, s *slot.Slot) bool {
	blockSlot := slot.NewFromUnixNano(block.Header.Timestamp)
	// The block corresponding to the current slot has already been generated.
	if slot.LessEqual(s, blockSlot) {
		return false
	}

	// Check whether the remaining time is enough until the next block
	// generation time.
	if !slot.IsNextTo(s, blockSlot) && !s.TimesUp() {
		return false
	}

	timeLeft := s.RemainingTimeMS()
	if timeLeft < 0 {
		logger.Debug().Int64("remaining time", timeLeft).Msg("no time left to produce block")
		return false
	}

	return true
}
