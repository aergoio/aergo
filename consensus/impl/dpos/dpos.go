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
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

const (
	// blockProducers is the number of block producers
	blockProducers = 23
)

var (
	logger = log.NewLogger("dpos")

	lastJob *slot.Slot
)

// DPoS is the main data structure of DPoS consensus
type DPoS struct {
	ID peer.ID
	*component.ComponentHub
	bpc  *bp.Cluster
	bf   *BlockFactory
	quit chan interface{}
}

// Status shows DPoS consensus's current status
type bpInfo struct {
	bestBlock *types.Block
	slot      *slot.Slot
}

// New returns a new DPos object
func New(cfg *config.Config, hub *component.ComponentHub) (consensus.Consensus, error) {
	Init(cfg.Consensus)

	bpc, err := bp.NewCluster(cfg.Consensus.BpIds, blockProducers)
	if err != nil {
		return nil, err
	}

	id, privKey := p2p.GetMyID()

	quitC := make(chan interface{})

	return &DPoS{
		ID:           id,
		ComponentHub: hub,
		bpc:          bpc,
		bf:           NewBlockFactory(hub, id, privKey, quitC),
		quit:         quitC,
	}, nil
}

// Init initilizes the DPoS parameters.
func Init(cfg *config.ConsensusConfig) {
	consensus.InitBlockInterval(cfg.BlockInterval)
	slot.Init(cfg.BlockInterval, blockProducers)
}

// Ticker returns a time.Ticker for the main consensus loop.
func (dpos *DPoS) Ticker() *time.Ticker {
	return time.NewTicker(consensus.BlockInterval / 100)

}

// QueueJob send a block triggering information to jq.
func (dpos *DPoS) QueueJob(now time.Time, jq chan<- interface{}) {
	bpi := dpos.getBpInfo(now, lastJob)
	if bpi != nil {
		jq <- bpi
		lastJob = bpi.slot
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

// IsBlockValid checks the DPoS consensus level validity of a block
func (dpos *DPoS) IsBlockValid(block *types.Block, bestBlock *types.Block) error {
	id, err := block.BpID()
	if err != nil {
		return &consensus.ErrorConsensus{Msg: "bad public key in block", Err: err}
	}

	if id == dpos.ID && block.PrevID() != bestBlock.ID() {
		return &consensus.ErrorConsensus{
			Msg: fmt.Sprintf(
				"reorganization occurred after block production: parent: %v (curr: %v), best block: %v",
				block.PrevID(), block.ID(), bestBlock.ID()),
		}
	}

	ns := block.GetHeader().GetTimestamp()
	idx, ok := dpos.bpc.BpID2Index(id)
	s := slot.NewFromUnixNano(ns)
	// Check whether the BP ID is one of the current BP members and its
	// corresponding BP index is consistent with the block timestamp.
	if !ok || !s.IsFor(idx) {
		return &consensus.ErrorConsensus{
			Msg: fmt.Sprintf("BP %v is not permitted for the time slot %v", block.ID(), time.Unix(0, ns)),
		}
	}

	valid, err := block.VerifySign()
	if !valid {
		return &consensus.ErrorConsensus{Msg: "bad block signature", Err: err}
	}

	return nil
}

// StatusUpdate updates the last irreversible block (LIB).
func (dpos *DPoS) StatusUpdate() {
}

func (dpos *DPoS) bpIdx() uint16 {
	idx, exist := dpos.bpc.BpID2Index(dpos.ID)
	if !exist {
		logger.Fatal().Str("id", enc.ToString([]byte(dpos.ID))).Msg("BP has no correct BP membership")
	}

	return idx
}

func (dpos *DPoS) getBpInfo(now time.Time, slotQueued *slot.Slot) *bpInfo {
	s := slot.Time(now)

	if !s.IsFor(dpos.bpIdx()) {
		return nil
	}

	// already queued slot.
	if slot.Equal(s, slotQueued) {
		return nil
	}

	block := chain.GetBestBlock(dpos)
	if block == nil {
		return nil
	}

	if !isBpTiming(block, s) {
		return nil
	}

	return &bpInfo{
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
