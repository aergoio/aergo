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
	consensus.ChainDB
	*component.ComponentHub
	bpc  *bp.Cluster
	bf   *BlockFactory
	quit chan interface{}
}

// Status shows DPoS consensus's current status
type bpInfo struct {
	consensus.ChainDB
	bestBlock *types.Block
	slot      *slot.Slot
}

func (bi *bpInfo) updateBestBLock() *types.Block {
	block, _ := bi.GetBestBlock()
	if block != nil {
		bi.bestBlock = block
	}

	return block
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub, cdb, sdb)
	}
}

// New returns a new DPos object
func New(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) (consensus.Consensus, error) {
	bpc, err := bp.NewCluster(cdb)
	if err != nil {
		return nil, err
	}

	Init(bpc.Size())

	quitC := make(chan interface{})

	return &DPoS{
		Status:       NewStatus(bpc, cdb, sdb, cfg.Blockchain.ForceResetHeight),
		ComponentHub: hub,
		ChainDB:      cdb,
		bpc:          bpc,
		bf:           NewBlockFactory(hub, sdb, quitC),
		quit:         quitC,
	}, nil
}

// Init initilizes the DPoS parameters.
func Init(bpCount uint16) {
	blockProducers = bpCount
	majorityCount = blockProducers*2/3 + 1
	// Collect voting for BPs during 10 rounds.
	initialBpElectionPeriod = types.BlockNo(blockProducers) * 10
	slot.Init(consensus.BlockIntervalSec, blockProducers)
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

// VerifySign reports the validity of the block signature.
func (dpos *DPoS) VerifySign(block *types.Block) error {
	valid, err := block.VerifySign()
	if !valid || err != nil {
		return &consensus.ErrorConsensus{Msg: "bad block signature", Err: err}
	}
	return nil
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

	block, _ := dpos.GetBestBlock()
	if block == nil {
		return nil
	}

	if !isBpTiming(block, s) {
		return nil
	}

	return &bpInfo{
		ChainDB:   dpos.ChainDB,
		bestBlock: block,
		slot:      s,
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
