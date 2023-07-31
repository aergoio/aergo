/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package dpos

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/impl/dpos/bp"
	"github.com/aergoio/aergo/v2/consensus/impl/dpos/slot"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

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

func (bi *bpInfo) updateBestBlock() *types.Block {
	block, _ := bi.GetBestBlock()
	if block != nil {
		bi.bestBlock = block
	}

	return block
}

type bfWork struct {
	context context.Context
	bpi     *bpInfo
}

// GetName returns the name of the consensus.
func GetName() string {
	return consensus.ConsensusName[consensus.ConsensusDPOS]
}

// GetConstructor build and returns consensus.Constructor from New function.
func GetConstructor(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) consensus.Constructor {
	return func() (consensus.Consensus, error) {
		return New(cfg, hub, cdb, sdb)
	}
}

func getStateDB(cfg *config.Config, cdb consensus.ChainDB, sdb *state.ChainStateDB) (*state.StateDB, error) {
	if cfg.Blockchain.VerifyOnly {
		vprInitBlockNo := func(blockNo types.BlockNo) types.BlockNo {
			if blockNo == 0 {
				return blockNo
			}
			return blockNo - 1
		}

		// Initialize the voting power ranking.
		if block, err := cdb.GetBlockByNo(vprInitBlockNo(cfg.Blockchain.VerifyBlock)); err != nil {
			return nil, err
		} else {
			return sdb.OpenNewStateDB(block.GetHeader().GetBlocksRootHash()), nil
		}
	}
	return sdb.GetStateDB(), nil
}

// New returns a new DPos object
func New(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDB,
	sdb *state.ChainStateDB) (consensus.Consensus, error) {

	chain.DecorateBlockRewardFn(sendVotingReward)

	bpc, err := bp.NewCluster(cdb)
	if err != nil {
		return nil, err
	}

	var state *state.StateDB
	if state, err = getStateDB(cfg, cdb, sdb); err != nil {
		return nil, err
	}

	if err = InitVPR(state); err != nil {
		return nil, err
	}

	Init(bpc.Size())

	quitC := make(chan interface{})

	return &DPoS{
		Status:       NewStatus(bpc, cdb, sdb, cfg.Blockchain.ForceResetHeight),
		ComponentHub: hub,
		ChainDB:      cdb,
		bpc:          bpc,
		bf:           NewBlockFactory(hub, sdb, quitC, cfg.Hardfork, cfg.Consensus.NoTimeoutTxEviction),
		quit:         quitC,
	}, nil
}

func sendVotingReward(bState *state.BlockState, dummy []byte) error {
	vrSeed := func(stateRoot []byte) int64 {
		return int64(binary.LittleEndian.Uint64(stateRoot))
	}

	vaultID := types.ToAccountID([]byte(types.AergoVault))
	vs, err := bState.GetAccountState(vaultID)
	if err != nil {
		logger.Info().Err(err).Msg("skip voting reward")
		return nil
	}

	vaultBalance := vs.GetBalanceBigInt()

	if vaultBalance.Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil
	}

	reward := system.GetVotingRewardAmount()
	if vaultBalance.Cmp(reward) < 0 {
		reward = new(big.Int).Set(vaultBalance)
	}

	addr, err := system.PickVotingRewardWinner(vrSeed(bState.PrevBlockHash()))
	if err != nil {
		logger.Debug().Err(err).Msg("no voting reward winner")
		return nil
	}

	ID := types.ToAccountID(addr)
	s, err := bState.GetAccountState(ID)
	if err != nil {
		logger.Info().Err(err).Msg("skip voting reward")
		return nil
	}

	newBalance := new(big.Int).Add(s.GetBalanceBigInt(), reward)
	s.Balance = newBalance.Bytes()

	err = bState.PutState(ID, s)
	if err != nil {
		return err
	}

	vs.Balance = vaultBalance.Sub(vaultBalance, reward).Bytes()
	if err = bState.PutState(vaultID, vs); err != nil {
		return err
	}

	bState.SetConsensus(addr)

	logger.Debug().
		Str("address", types.EncodeAddress(addr)).
		Str("amount", reward.String()).
		Str("new balance", newBalance.String()).
		Str("vault balance", vaultBalance.String()).
		Msg("voting reward winner appointed")

	return nil
}

func InitVPR(sdb *state.StateDB) error {
	s, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return err
	}
	return system.InitVotingPowerRank(s)
}

// Init initializes the DPoS parameters.
func Init(bpCount uint16) {
	blockProducers = bpCount
	majorityCount = blockProducers*2/3 + 1
	// Collect voting for BPs during 10 rounds.
	initialBpElectionPeriod = types.BlockNo(blockProducers) * 10
	slot.Init(consensus.BlockIntervalSec)
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

func (dpos *DPoS) GetType() consensus.ConsensusType {
	return consensus.ConsensusDPOS
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

func (dpos *DPoS) bpid() types.PeerID {
	return p2pkey.NodeID()
}

// VerifyTimestamp checks the validity of the block timestamp.
func (dpos *DPoS) VerifyTimestamp(block *types.Block) bool {

	if ts := block.GetHeader().GetTimestamp(); slot.NewFromUnixNano(ts).IsFuture() {
		logger.Error().Str("BP", block.BPID2Str()).Str("id", block.ID()).
			Time("timestamp", time.Unix(0, ts)).Msg("block has a future timestamp")
		return false
	}

	// Reject the blocks with no <= LIB since it cannot lead to a
	// reorganization.
	if dpos.Status != nil && block.BlockNo() <= dpos.libNo() {
		logger.Error().Str("BP", block.BPID2Str()).Str("id", block.ID()).
			Uint64("block no", block.BlockNo()).Uint64("lib no", dpos.libNo()).
			Msg("too small block number (<= LIB number)")
		return false
	}

	return true
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

	idx := dpos.bpc.BpID2Index(id)
	ns := block.GetHeader().GetTimestamp()
	s := slot.NewFromUnixNano(ns)
	// Check whether the BP ID is one of the current BP members and its
	// corresponding BP index is consistent with the block timestamp.
	if !s.IsFor(idx, dpos.bpc.Size()) {
		return &consensus.ErrorConsensus{
			Msg: fmt.Sprintf("BP %v (idx: %v) is not permitted for the time slot %v (%v)",
				block.BPID2Str(), idx, time.Unix(0, ns), s.NextBpIndex(dpos.bpc.Size())),
		}
	}

	return nil
}

func (dpos *DPoS) bpIdx() bp.Index {
	return dpos.bpc.BpID2Index(dpos.bpid())
}

func (dpos *DPoS) getBpInfo(now time.Time) *bpInfo {
	s := slot.Time(now)

	if !s.IsFor(dpos.bpIdx(), dpos.bpc.Size()) {
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

// ConsensusInfo returns the basic DPoS-related info.
func (dpos *DPoS) ConsensusInfo() *types.ConsensusInfo {
	withLock := func(fn func()) {
		dpos.RLock()
		defer dpos.RUnlock()
		fn()
	}

	ci := &types.ConsensusInfo{Type: GetName()}
	withLock(func() {
		ci.Bps = dpos.bpc.BPs()

	})

	if dpos.done {
		var lpbNo types.BlockNo

		withLock(func() {
			lpbNo = dpos.lpbNo()
		})

		if lpbNo > 0 {
			if block, err := dpos.GetBlockByNo(lpbNo); err == nil {
				type lpbInfo struct {
					BPID      string
					Height    types.BlockNo
					Hash      string
					Timestamp string
				}
				s := struct {
					NodeID              string
					RecentBlockProduced lpbInfo
				}{
					NodeID: dpos.bf.ID,
					RecentBlockProduced: lpbInfo{
						BPID:      block.BPID2Str(),
						Height:    lpbNo,
						Hash:      block.ID(),
						Timestamp: block.Localtime().String(),
					},
				}
				if m, err := json.Marshal(s); err == nil {
					ci.Info = string(m)
				}
			}
		}
	}

	return ci
}

var dummyRaft consensus.DummyRaftAccessor

func (dpos *DPoS) RaftAccessor() consensus.AergoRaftAccessor {
	return &dummyRaft
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

func (dpos *DPoS) NeedNotify() bool {
	return true
}

func (dpos *DPoS) HasWAL() bool {
	return false
}

func (dpos *DPoS) IsForkEnable() bool {
	return true
}

func (dpos *DPoS) IsConnectedBlock(block *types.Block) bool {
	_, err := dpos.ChainDB.GetBlock(block.BlockHash())
	if err == nil {
		return true
	}

	return false
}

func (dpos *DPoS) ConfChange(req *types.MembershipChange) (*consensus.Member, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (dpos *DPoS) ConfChangeInfo(requestID uint64) (*types.ConfChangeProgress, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (dpos *DPoS) MakeConfChangeProposal(req *types.MembershipChange) (*consensus.ConfChangePropose, error) {
	return nil, consensus.ErrNotSupportedMethod
}

func (dpos *DPoS) ClusterInfo(bestBlockHash []byte) *types.GetClusterInfoResponse {
	return &types.GetClusterInfoResponse{ChainID: nil, Error: consensus.ErrNotSupportedMethod.Error(), MbrAttrs: nil, HardStateInfo: nil}
}

func ValidateGenesis(genesis *types.Genesis) error {
	return nil
}
