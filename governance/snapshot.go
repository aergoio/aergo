package governance

import (
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/governance/enterprise"
	"github.com/aergoio/aergo/v2/governance/name"
	"github.com/aergoio/aergo/v2/governance/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type Snapshot struct {
	// do not access member directly
	cfg             *Config
	ctx             *ChainContext
	systemParams    *system.Parameters
	votingPowerRank *system.Vpr
	nameParams      *name.Names
}

func (ss *Snapshot) Init(ctx *ChainContext, getter system.DataGetter) error {
	var err error

	// load ctx
	ss.ctx = ctx

	// load system
	ss.systemParams = system.NewParameters()

	// system.LoadParam(getter)
	ss.votingPowerRank, err = system.LoadVpr(getter)
	if err != nil {
		return err
	}

	// load name
	ss.nameParams = name.NewNames(ss.systemParams.GetGasPrice())

	return nil
}

func (ss *Snapshot) Copy() *Snapshot {
	new := &Snapshot{
		systemParams: nil,
		nameParams:   nil,
	}
	// TODO
	// for k, v := range ss.systemParams {
	// new.systemParams[k] = big.NewInt(0).Set(v)
	// }
	return new
}

// name
func (ss *Snapshot) GetNameOwner(scs *state.ContractState, account []byte) []byte {

	return name.GetOwner(scs, account) // 1.메모리 // 2. state db / 3. trie (real db ) 4. default value
}

func (ss *Snapshot) GetNameAddress(scs *state.ContractState, account []byte) []byte {
	return name.GetAddress(scs, account)
}

// system
func (ss *Snapshot) GetSystemBpCount() int {
	// get from memory
	param := ss.systemParams.GetBpCount()
	if param != nil {
		return int(param.Int64())
	}

	// get from state
	param = system.GetBpCountFromState(ss.ctx.scs)
	if param != nil {
		return int(param.Int64())
	}

	// TODO : return default value
	return 0
}

func (ss *Snapshot) GetSystemStakingMinimum() *big.Int {
	// get from memory
	param := ss.systemParams.GetStakingMinimum()
	if param != nil {
		return param
	}

	// get from state
	param = system.GetStakingMinimumFromState(ss.ctx.scs)
	if param != nil {
		return param
	}

	// TODO : return default value
	return nil
}

func (ss *Snapshot) GetSystemNamePrice() *big.Int {
	// get from memory
	param := ss.systemParams.GetNamePrice()
	if param != nil {
		return param
	}

	// get from state
	param = system.GetNamePriceFromState(ss.ctx.scs)
	if param != nil {
		return param
	}

	// TODO : return default value
	return nil
}

func (ss *Snapshot) GetSystemGasPrice() *big.Int {
	// get from memory
	param := ss.systemParams.GetGasPrice()
	if param != nil {
		return param
	}

	// get from state
	param = system.GetNamePriceFromState(ss.ctx.scs)
	if param != nil {
		return param
	}

	// TODO : return default value
	return nil
}

// voting
func (ss *Snapshot) GetTotalVotingPower() *big.Int {
	return ss.votingPowerRank.GetTotalVotingPower()
}

func (ss *Snapshot) GetStakingTotal() *big.Int {
	return ss.systemParams.GetStakingMinimum()
}

func (ss *Snapshot) GetVotingRewardAmount() *big.Int {
	return nil
}

func (ss *Snapshot) PickVotingRewardWinner(seed int64) (types.Address, error) {
	return ss.votingPowerRank.PickVotingRewardWinner(seed)
}

// enterprise
func (ss *Snapshot) GetEnterpriseConfWhiteList(r enterprise.AccountStateReader) (*types.EnterpriseConfig, error) {
	return enterprise.GetConf(r, enterprise.AccountWhite)
}

func (ss *Snapshot) Execute(ccc consensus.ChainConsensusCluster, ctx *ChainContext) ([]*types.Event, error) {
	var err error

	// 고민 : ctx 도 prev / next 를 나눌 필요가 있을까?

	var events []*types.Event
	switch ctx.governance {
	case types.AergoSystem:
		events, err = system.ExecuteSystemTx(ctx.scs, ctx.txInfo, ctx.Sender, ctx.Receiver, ctx.blockInfo)
	case types.AergoName:
		events, err = name.ExecuteNameTx(ctx.bs, ctx.scs, ctx.txInfo, ctx.Sender, ctx.Receiver, ctx.blockInfo)
	case types.AergoEnterprise:
		events, err = enterprise.ExecuteEnterpriseTx(ctx.bs, ccc, ctx.scs, ctx.txInfo, ctx.Sender, ctx.Receiver, ctx.blockInfo.No)
		if err != nil {
			err = NewGovEntErr(err)
		}
	default:
		// ss.log.Warn().Str("governance", ctx.governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}
	if err == nil {
		err = ctx.bs.StateDB.StageContractState(ctx.scs)
	}

	return events, err
}

func (ss *Snapshot) ValidateMempool(scs *state.ContractState, sdb *state.StateDB, account []byte) error {
	switch string(ss.ctx.txInfo.GetRecipient()) {
	case types.AergoSystem:
		sender, err := sdb.GetAccountStateV(account)
		if err != nil {
			return err
		}
		nextBlockInfo := types.BlockHeaderInfo{
			No:          ss.ctx.bestBlockNo + 1,
			ForkVersion: ss.ctx.nextBlockVersion,
		}
		if _, err := system.ValidateSystemTx(ss.cfg.proposals, account, ss.ctx.txInfo, sender, scs, &nextBlockInfo); err != nil {
			return err
		}
	case types.AergoName:
		systemcs, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
		if err != nil {
			return err
		}
		sender, err := sdb.GetAccountStateV(account)
		if err != nil {
			return err
		}
		if _, err := name.ValidateNameTx(ss.ctx.txInfo, sender, scs, systemcs); err != nil {
			return err
		}
	case types.AergoEnterprise:
		enterprisecs, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte(types.AergoEnterprise)))
		if err != nil {
			return err
		}
		sender, err := sdb.GetAccountStateV(account)
		if err != nil {
			return err
		}
		if _, err := enterprise.ValidateEnterpriseTx(ss.ctx.txInfo, sender, enterprisecs, ss.ctx.bestBlockNo+1); err != nil {
			return err
		}
	}
	return nil
}
