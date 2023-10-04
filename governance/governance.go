package governance

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/governance/enterprise"
	"github.com/aergoio/aergo/v2/governance/name"
	"github.com/aergoio/aergo/v2/governance/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

func NewGovernance(genesis *types.Genesis, consensus string, state system.DataGetter) *Governance {
	gov := &Governance{}
	gov.log = log.NewLogger("system")
	gov.cfg = NewConfig(genesis, consensus)
	return gov
}

type Governance struct {
	mtx sync.Mutex
	log *log.Logger

	// immutable params
	cfg *Config
	ctx *ChainContext

	// snapshot params
	// init : update beforeExec, afterExec to init
	// commit : update beforeExec to afterExec
	// revert : update afterExec to beforeExec
	// reorg : update beforeExec, afterExec to db(reorg)
	initial    *Snapshot
	beforeExec *Snapshot
	afterExec  *Snapshot
}

// prev 를 보는 상황 -> tx Execute 를 제외한 다른 모듈들이 조회할 때
// next 를 보는 상황 -> tx execute 에서 업데이트할 때

func (g *Governance) Snapshot() *Snapshot {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	return g.afterExec.Copy()
}

func (g *Governance) Commit() {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	g.beforeExec = g.afterExec.Copy()
}

func (g *Governance) Revert() {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	g.afterExec = g.beforeExec.Copy()
}

func (g *Governance) Reorg() {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	var Reorg *Snapshot
	// TODO : init reorg from db

	g.beforeExec = Reorg.Copy()
	g.afterExec = Reorg.Copy()
}

func NewChainContext(blockInfo *types.BlockHeaderInfo, txHash []byte, txBody *types.TxBody, bs *state.BlockState, sender, receiver *state.V) (*ChainContext, error) {
	scs, err := bs.StateDB.OpenContractState(receiver.AccountID(), receiver.State())
	if err != nil {
		return nil, err
	}
	govName := string(txBody.Recipient)
	return &ChainContext{
		blockInfo: blockInfo,
		txHash:    txHash,
		txInfo:    txBody,
		govName:   govName,
		bs:        bs,
		scs:       scs,
		Sender:    sender,
		Receiver:  receiver,
	}, nil
}

type ChainContext struct {
	bestBlockNo      types.BlockNo
	nextBlockVersion int32

	blockInfo *types.BlockHeaderInfo
	txHash    []byte
	txInfo    *types.TxBody
	callInfo  *types.CallInfo

	// state per tx
	govName  string
	bs       *state.BlockState
	scs      *state.ContractState
	Sender   *state.V
	Receiver *state.V
}

func (g *Governance) GetSystemValue(key types.SystemValue) (*big.Int, error) {
	switch key {
	case types.StakingTotal:
		return g.beforeExec.GetStakingTotal(), nil
	case types.StakingMin:
		return g.beforeExec.GetSystemStakingMinimum(), nil
	case types.GasPrice:
		return g.beforeExec.GetSystemGasPrice(), nil
	case types.NamePrice:
		return g.beforeExec.GetSystemNamePrice(), nil
	case types.TotalVotingPower:
		return g.beforeExec.GetTotalVotingPower(), nil
	case types.VotingReward:
		return g.beforeExec.GetVotingRewardAmount(), nil
	}
	return nil, fmt.Errorf("unsupported system value : %s", key)
}

func (g *Governance) GetVotes(id string, n uint32) (*types.VoteList, error) {
	if g.cfg.consensusType != consensus.ConsensusName[consensus.ConsensusDPOS] {
		return nil, ErrNotSupportedConsensus
	}
	return g.beforeExec.GetVoteList([]byte(id), n)
}

func (g *Governance) GetAccountVote(addr []byte) (*types.AccountVoteInfo, error) {
	if g.cfg.consensusType != consensus.ConsensusName[consensus.ConsensusDPOS] {
		return nil, ErrNotSupportedConsensus
	}

	voteInfo, err := g.beforeExec.GetVoteList(addr)
	if err != nil {
		return nil, err
	}
	return &types.AccountVoteInfo{Voting: voteInfo}, nil

}

func (g *Governance) GetStaking(addr []byte) (*types.Staking, error) {
	if g.cfg.consensusType != consensus.ConsensusName[consensus.ConsensusDPOS] {
		return nil, ErrNotSupportedConsensus
	}

	sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	scs, err := sdb.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	namescs, err := sdb.GetNameAccountState()
	if err != nil {
		return nil, err
	}
	staking, err := system.GetStaking(scs, name.GetAddressFromState(namescs, addr))
	if err != nil {
		return nil, err
	}
	return staking, nil
}

func (g *Governance) GetNameInfo(qname string, blockNo types.BlockNo) (*types.NameInfo, error) {

	return name.GetNameInfo(stateDB, qname)
}

func (g *Governance) GetEnterpriseConf(key string) (*types.EnterpriseConfig, error) {
	// sdb := cs.sdb.OpenNewStateDB(cs.sdb.GetRoot())
	// if strings.ToUpper(key) != enterprise.AdminsKey {
	// 	return enterprise.GetConf(sdb, key)
	// }
	return enterprise.GetAdmin(sdb)
}
