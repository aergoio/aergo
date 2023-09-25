/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestBasicStakingUnstaking(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	minplusmin := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
	sender.AddBalance(minplusmin)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	staking, err := newSysCmd(sender.ID(), tx.GetBody(), sender, receiver, scs, blockInfo)
	event, err := staking.run()
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	assert.Equal(t, event.EventName, "stake", "event name")
	assert.Equal(t, event.JsonArgs, "{\"who\":\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"amount\":\"10000000000000000000000\"}", "event args")
	saved, err := getStaking(scs, tx.Body.Account)
	assert.Equal(t, types.StakingMinimum.Bytes(), saved.Amount, "saved staking value")
	total, err := getStakingTotal(scs)
	assert.Equal(t, types.StakingMinimum, total, "total value")
	blockInfo.No += (StakingDelay - 1)
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, blockInfo)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	blockInfo.No++
	unstake, err := newSysCmd(sender.ID(), tx.GetBody(), sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "should be success")
	event, err = unstake.run()
	assert.NoError(t, err, "should be success")
	assert.Equal(t, sender.Balance(), minplusmin, "sender.Balance() cacluation failed")
	assert.Equal(t, event.EventName, "unstake", "event name")
	assert.Equal(t, event.JsonArgs, "{\"who\":\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"amount\":\"10000000000000000000000\"}", "event args")
	saved, err = getStaking(scs, tx.Body.Account)
	assert.Equal(t, new(big.Int).SetUint64(0).Bytes(), saved.Amount, "saved staking value")
	total, err = getStakingTotal(scs)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, new(big.Int).SetUint64(0), total, "total value")
}

func TestBasicStakingUnstakingV2(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	minplusmin := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
	sender.AddBalance(minplusmin)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0), ForkVersion: 2}
	staking, err := newSysCmd(sender.ID(), tx.GetBody(), sender, receiver, scs, blockInfo)
	event, err := staking.run()
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	assert.Equal(t, event.EventName, "stake", "event name")
	assert.Equal(t, event.JsonArgs, "[\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", {\"_bignum\":\"10000000000000000000000\"}]", "event args")
	saved, err := getStaking(scs, tx.Body.Account)
	assert.Equal(t, types.StakingMinimum.Bytes(), saved.Amount, "saved staking value")
	total, err := getStakingTotal(scs)
	assert.Equal(t, types.StakingMinimum, total, "total value")
	blockInfo.No += (StakingDelay - 1)
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, blockInfo)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	blockInfo.No++
	unstake, err := newSysCmd(sender.ID(), tx.GetBody(), sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "should be success")
	event, err = unstake.run()
	assert.NoError(t, err, "should be success")
	assert.Equal(t, sender.Balance(), minplusmin, "sender.Balance() cacluation failed")
	assert.Equal(t, event.EventName, "unstake", "event name")
	assert.Equal(t, event.JsonArgs, "[\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", {\"_bignum\":\"10000000000000000000000\"}]", "event args")
	saved, err = getStaking(scs, tx.Body.Account)
	assert.Equal(t, new(big.Int).SetUint64(0).Bytes(), saved.Amount, "saved staking value")
	total, err = getStakingTotal(scs)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, new(big.Int).SetUint64(0), total, "total value")
}

func TestStaking1Unstaking2(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	tx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	sender.AddBalance(types.MaxAER)
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}

	stake, err := newSysCmd(sender.ID(), tx.GetBody(), sender, receiver, scs, blockInfo)
	_, err = stake.run()
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be 'MaxAER - StakingMin' after staking")

	blockInfo.No += (StakingDelay - 1)
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, blockInfo)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	blockInfo.No++
	tx.Body.Amount = new(big.Int).Add(types.StakingMinimum, types.StakingMinimum).Bytes()
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, blockInfo)
	assert.Error(t, err, "should return exceed error")
}

func TestUnstakingError(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.Equal(t, err, nil, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.Equal(t, err, nil, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte(`{"Name":"v1unstake"}`),
		},
	}
	sender, err := bs.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := bs.GetAccountStateV(tx.Body.Recipient)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.MaxAER)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	_, err = ExecuteSystemTx(scs, tx.Body, sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "should be success")
}
