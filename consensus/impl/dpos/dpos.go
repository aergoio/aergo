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
	peer "github.com/libp2p/go-libp2p-peer"
)

// DefaultDposBpNumber is the default number of block producers.
const DefaultDposBpNumber = 23

var (
	logger = log.NewLogger("dpos")

	// blockProducers is the number of block producers
	blockProducers          uint16
	majorityCount           uint16
	initialBpElectionPeriod types.BlockNo

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

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.ConsensusConfig, hub *component.ComponentHub, cdb consensus.ChainDbReader) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub, cdb)
	}
}

// New returns a new DPos object
func New(cfg *config.ConsensusConfig, hub *component.ComponentHub, cdb consensus.ChainDbReader) (consensus.Consensus, error) {
	bpc, err := bp.NewCluster(cfg, cdb)
	if err != nil {
		return nil, err
	}

	Init(bpc.Size(), cfg.BlockInterval)

	quitC := make(chan interface{})

	return &DPoS{
		Status:       NewStatus(bpc.Size(), cdb),
		ComponentHub: hub,
		bpc:          bpc,
		bf:           NewBlockFactory(hub, quitC),
		quit:         quitC,
	}, nil
}

// Init initilizes the DPoS parameters.
func Init(bpCount uint16, blockInterval int64) {
	blockProducers = bpCount
	majorityCount = blockProducers*2/3 + 1
	// Collect voting for BPs during 10 rounds.
	initialBpElectionPeriod = types.BlockNo(blockProducers) * 10
	slot.Init(blockInterval, blockProducers)
}

func consensusBlockCount(bpCount uint16) uint16 {
	return bpCount*2/3 + 1
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
	dpos.Status.setStateDB(sdb)
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
