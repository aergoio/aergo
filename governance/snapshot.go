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
	cfg *Config
	ctx *ChainContext

	// do not access directly
	systemParams    *system.Parameters
	nameParams      *name.Names
	votingPowerRank *system.Vpr

	// TODO : vote, staking, enterprise
}

func (ss *Snapshot) Init(cfg *Config, ctx *ChainContext, getter system.DataGetter) error {
	var err error

	// init cfg, ctx
	ss.cfg = cfg
	ss.ctx = ctx

	// load systems
	ss.systemParams = system.NewParameters()

	// load names
	ss.nameParams = name.NewNames()

	// load votingPowerRank
	ss.votingPowerRank, err = system.LoadVpr(getter)
	if err != nil {
		return err
	}

	return nil
}

func (ss *Snapshot) Copy() *Snapshot {
	return &Snapshot{
		cfg:          ss.cfg,
		ctx:          ss.ctx,
		systemParams: ss.systemParams.Copy(),
		nameParams:   ss.nameParams.Copy(),
		// TODO : copy voting
	}
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
	param = system.GetGasPriceFromState(ss.ctx.scs)
	if param != nil {
		return param
	}

	// TODO : return default value
	return nil
}

// name
func (ss *Snapshot) GetNameOwner(scs *state.ContractState, account []byte) []byte {
	// get from memory
	owner := ss.nameParams.GetOwner(account)
	if owner != nil {
		return owner
	}
	// get from state
	owner = name.GetOwnerFromState(scs, account)
	if owner != nil {
		return owner
	}

	// get default - nil
	return nil
}

func (ss *Snapshot) GetNameAddress(scs *state.ContractState, account []byte) []byte {
	// get from memory
	addr := ss.nameParams.GetAddress(account)
	if addr != nil {
		return addr
	}
	// get from state
	addr = name.GetAddressFromState(scs, account)
	if addr != nil {
		return addr
	}

	// get default - nil
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

func (ss *Snapshot) Execute(ccc consensus.ChainConsensusCluster) ([]*types.Event, error) {
	var err error

	var events []*types.Event
	switch ss.ctx.governance {
	case types.AergoSystem:
		events, err = system.ExecuteSystemTx(ss.cfg.proposals, ss.ctx.scs, ss.ctx.txInfo, ss.ctx.Sender, ss.ctx.Receiver, ss.ctx.blockInfo)
	case types.AergoName:
		events, err = name.ExecuteNameTx(ss.ctx.bs, ss.ctx.scs, ss.ctx.txInfo, ss.ctx.Sender, ss.ctx.Receiver, ss.ctx.blockInfo)
	case types.AergoEnterprise:
		events, err = enterprise.ExecuteEnterpriseTx(ss.ctx.bs, ccc, ss.ctx.scs, ss.ctx.txInfo, ss.ctx.Sender, ss.ctx.Receiver, ss.ctx.blockInfo.No)
		if err != nil {
			err = NewGovEntErr(err)
		}
	default:
		// ss.log.Warn().Str("governance", ctx.governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}
	if err == nil {
		err = ss.ctx.bs.StateDB.StageContractState(ss.ctx.scs)
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
