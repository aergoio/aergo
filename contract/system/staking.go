/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var consensusType string

var (
	stakingKey      = []byte("staking")
	stakingTotalKey = []byte("stakingtotal")

	ErrInvalidCandidate = errors.New("invalid candidate")
)

const StakingDelay = 60 * 60 * 24 //block interval
//const StakingDelay = 5

func InitGovernance(consensus string) {
	consensusType = consensus
}

type stakeCmd struct {
	*SystemContext
	amount *big.Int
}

func newStakeCmd(ctx *SystemContext) (sysCmd, error) {
	var (
		cmd    = &stakeCmd{SystemContext: ctx, amount: ctx.txBody.GetAmountBigInt()}
		staked = cmd.Staked
	)

	staked.Add(cmd.amount)
	staked.SetWhen(cmd.BlockInfo.No)

	return cmd, nil
}

func (c *stakeCmd) run() (*types.Event, error) {
	var (
		amount   = c.amount
		sender   = c.Sender
		receiver = c.Receiver
	)

	if err := c.updateStaking(); err != nil {
		return nil, err
	}
	if err := addTotal(c.scs, amount); err != nil {
		return nil, err
	}
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	if c.SystemContext.BlockInfo.ForkVersion < 2 {
		return &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "stake",
			JsonArgs: `{"who":"` +
				types.EncodeAddress(sender.ID()) +
				`", "amount":"` + amount.String() + `"}`,
		}, nil
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "stake",
		JsonArgs: `["` +
			types.EncodeAddress(sender.ID()) +
			`", {"_bignum":"` + amount.String() + `"}]`,
	}, nil
}

type unstakeCmd struct {
	*SystemContext
}

func newUnstakeCmd(ctx *SystemContext) (sysCmd, error) {
	return &unstakeCmd{
		SystemContext: ctx,
	}, nil
}

func (c *unstakeCmd) run() (*types.Event, error) {
	var (
		scs               = c.scs
		staked            = c.Staked
		sender            = c.Sender
		receiver          = c.Receiver
		balanceAdjustment = staked.Sub(c.txBody.GetAmountBigInt())
	)

	//blockNo will be updated in voting
	staked.SetWhen(c.BlockInfo.No)

	if err := c.updateStaking(); err != nil {
		return nil, err
	}
	if err := refreshAllVote(c.SystemContext); err != nil {
		return nil, err
	}
	if err := subTotal(scs, balanceAdjustment); err != nil {
		return nil, err
	}
	sender.AddBalance(balanceAdjustment)
	receiver.SubBalance(balanceAdjustment)
	if c.SystemContext.BlockInfo.ForkVersion < 2 {
		return &types.Event{
			ContractAddress: receiver.ID(),
			EventIdx:        0,
			EventName:       "unstake",
			JsonArgs: `{"who":"` +
				types.EncodeAddress(sender.ID()) +
				`", "amount":"` + balanceAdjustment.String() + `"}`,
		}, nil
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       "unstake",
		JsonArgs: `["` +
			types.EncodeAddress(sender.ID()) +
			`", {"_bignum":"` + balanceAdjustment.String() + `"}]`,
	}, nil
}

func setStaking(scs *state.ContractState, who []byte, staking *types.Staking) error {
	key := append(stakingKey, who...)
	return scs.SetData(key, serializeStaking(staking))
}

func getStaking(scs *state.ContractState, who []byte) (*types.Staking, error) {
	key := append(stakingKey, who...)
	data, err := scs.GetData(key)
	if err != nil {
		return nil, err
	}
	var staking types.Staking
	if len(data) != 0 {
		return deserializeStaking(data), nil
	}
	return &staking, nil
}

func GetStaking(scs *state.ContractState, address []byte) (*types.Staking, error) {
	if address != nil {
		return getStaking(scs, address)
	}
	return nil, errors.New("invalid argument: address should not be nil")
}

func GetStakingTotal(ar AccountStateReader) (*big.Int, error) {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	return getStakingTotal(scs)
}

func getStakingTotal(scs *state.ContractState) (*big.Int, error) {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func addTotal(scs *state.ContractState, amount *big.Int) error {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return err
	}
	total := new(big.Int).SetBytes(data)
	return scs.SetData(stakingTotalKey, new(big.Int).Add(total, amount).Bytes())
}

func subTotal(scs *state.ContractState, amount *big.Int) error {
	data, err := scs.GetData(stakingTotalKey)
	if err != nil {
		return err
	}
	total := new(big.Int).SetBytes(data)
	return scs.SetData(stakingTotalKey, new(big.Int).Sub(total, amount).Bytes())
}

func serializeStaking(v *types.Staking) []byte {
	var ret []byte
	if v != nil {
		when := make([]byte, 8)
		binary.LittleEndian.PutUint64(when, v.GetWhen())
		ret = append(ret, when...)
		ret = append(ret, v.GetAmount()...)
	}
	return ret
}

func deserializeStaking(data []byte) *types.Staking {
	when := binary.LittleEndian.Uint64(data[:8])
	amount := data[8:]
	return &types.Staking{Amount: amount, When: when}
}
