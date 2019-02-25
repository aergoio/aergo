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
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"

	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	minplusmin := new(big.Int).Add(types.StakingMinimum, types.StakingMinimum)
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV(tx.Body.Recipient)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(minplusmin)

	_, err = staking(tx.Body, sender, receiver, scs, 0)
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	saved, err := getStaking(scs, tx.Body.Account)
	assert.Equal(t, types.StakingMinimum.Bytes(), saved.Amount, "saved staking value")

	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	ci, err := ValidateSystemTx(account, tx.GetBody(), nil, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	ci, err = ValidateSystemTx(account, tx.GetBody(), nil, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	_, err = unstaking(tx.Body, sender, receiver, scs, StakingDelay, ci)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, sender.Balance(), minplusmin, "sender.Balance() cacluation failed")
}

func TestStaking1Unstaking2(t *testing.T) {
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
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV(tx.Body.Recipient)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.MaxAER)

	_, err = staking(tx.Body, sender, receiver, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be 'MaxAER - StakingMin' after staking")

	tx.Body.Amount = new(big.Int).Add(types.StakingMinimum, types.StakingMinimum).Bytes()
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	ci, err := ValidateSystemTx(account, tx.GetBody(), sender, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")

	ci, err = ValidateSystemTx(account, tx.GetBody(), sender, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	_, err = unstaking(tx.Body, sender, receiver, scs, StakingDelay, ci)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, sender.Balance().Bytes(), types.MaxAER.Bytes(), "sender.Balance() cacluation failed")
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
