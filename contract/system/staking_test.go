/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/types"
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

	ci, err := ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, 0)
	_, err = staking(tx.Body, sender, receiver, scs, 0, ci)
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	saved, err := getStaking(scs, tx.Body.Account)
	assert.Equal(t, types.StakingMinimum.Bytes(), saved.Amount, "saved staking value")
	total, err := getStakingTotal(scs)
	assert.Equal(t, types.StakingMinimum, total, "total value")

	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	ci, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	ci, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	_, err = unstaking(tx.Body, sender, receiver, scs, StakingDelay, ci)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, sender.Balance(), minplusmin, "sender.Balance() cacluation failed")
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

	ci, err := ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, 0)
	_, err = staking(tx.Body, sender, receiver, scs, 0, ci)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be 'MaxAER - StakingMin' after staking")

	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	tx.Body.Amount = new(big.Int).Add(types.StakingMinimum, types.StakingMinimum).Bytes()
	_, err = ValidateSystemTx(sender.ID(), tx.GetBody(), sender, scs, StakingDelay)
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
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV(tx.Body.Recipient)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.MaxAER)

	_, err = ExecuteSystemTx(scs, tx.Body, sender, receiver, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "should be success")
}
