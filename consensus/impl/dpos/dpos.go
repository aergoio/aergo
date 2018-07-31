/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"fmt"
	"time"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/consensus/impl/dpos/param"
	"github.com/aergoio/aergo/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/consensus/util"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

var (
	logger  *log.Logger
	lastJob *slot.Slot
)

// DPoS is the main data structure of DPoS consensus
type DPoS struct {
	ID peer.ID
	*component.ComponentHub
	bpc            *bp.Cluster
	bf             *BlockFactory
	onReorganizing util.BcReorgStatus
}

// Status shows DPoS consensus's current status
type bpInfo struct {
	bestBlock *types.Block
	slot      *slot.Slot
}

func init() {
	logger = log.NewLogger(log.DPOS)
}

// New returns a new DPos object
func New(cfg *config.Config, hub *component.ComponentHub) (consensus.Consensus, error) {
	param.Init(cfg.Consensus.BlockInterval)

	bpc, err := bp.NewCluster(cfg.Consensus.BpIds)
	if err != nil {
		return nil, err
	}
	id, privKey := p2p.GetMyID()

	return &DPoS{
		ID:             id,
		ComponentHub:   hub,
		bpc:            bpc,
		bf:             NewBlockFactory(hub, id, privKey),
		onReorganizing: util.BcNoReorganizing,
	}, nil
}

// Ticker returns a time.Ticker for the main consensus loop.
func (dpos *DPoS) Ticker() *time.Ticker {
	return time.NewTicker(param.LoopInterval)
}

// QueueJob send a block triggering infomation to jq.
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

// IsBlockValid checks the DPoS consensus level validity of a block
func (dpos *DPoS) IsBlockValid(block *types.Block) error {
	id, err := block.BpID()
	if err != nil {
		return &consensus.ErrorConsensus{Msg: "bad public key in block", Err: err}
	}

	if !dpos.bpc.Has(id) {
		return &consensus.ErrorConsensus{
			Msg: fmt.Sprintf("BP %v not in an BP cluster", id.Pretty()),
		}
	}

	valid, err := block.VerifySign()
	if !valid {
		return &consensus.ErrorConsensus{Msg: "bad block signature", Err: err}
	}

	return nil
}

// IsBlockReorganizing reports whether the blockchain is currently under
// reorganization.
func (dpos *DPoS) IsBlockReorganizing() bool {
	return util.OnReorganizing(&dpos.onReorganizing)
}

// SetReorganizing sets dpos.onReorganizing to 'OnReorganization.'
func (dpos *DPoS) SetReorganizing() {
	util.SetReorganizing(&dpos.onReorganizing)
}

// SetReorganizing sets dpos.onReorganizing to 'NoReorganization.'
func (dpos *DPoS) UnsetReorganizing() {
	util.UnsetReorganizing(&dpos.onReorganizing)
}

func (dpos *DPoS) bpIdx() uint16 {
	idx, exist := dpos.bpc.BpID2Index(dpos.ID)
	if !exist {
		logger.Fatalf("BP %s has no correct BP membership", dpos.ID.Pretty())
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

	block := util.GetBestBlock(dpos)
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
	blockSlot := slot.Unix(block.Header.Timestamp)
	if slot.LessEqual(s, blockSlot) {
		return false
	}

	if !slot.IsNextTo(s, blockSlot) && s.RemainingTimeMS() > param.BpMinTimeLimitMs {
		return false
	}

	return true
}
