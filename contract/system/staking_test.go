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

	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'s'},
		},
	}
	senderState := &types.State{Balance: types.StakingMinimum.Bytes()}
	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalanceBigInt().Bytes(), new(big.Int).SetUint64(0).Bytes(), "sender.GetBalanceBigInt() should be 0 after staking")

	tx.Body.Payload = []byte{'u'}
	err = unstaking(tx.Body, senderState, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalanceBigInt().Uint64(), uint64(0), "sender.GetBalanceBigInt() should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, senderState.GetBalanceBigInt().Bytes(), types.StakingMinimum.Bytes(), "sender.GetBalanceBigInt() cacluation failed")
}

func TestStaking1Unstaking2(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.Equal(t, err, nil, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.Equal(t, err, nil, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'s'},
		},
	}
	senderState := &types.State{Balance: types.MaxAER.Bytes()}
	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalanceBigInt().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.GetBalanceBigInt() should be 'MaxAER - StakingMin' after staking")

	tx.Body.Amount = new(big.Int).Add(types.StakingMinimum, types.StakingMinimum).Bytes()
	tx.Body.Payload = []byte{'u'}
	err = unstaking(tx.Body, senderState, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalanceBigInt().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.GetBalanceBigInt() should not be changed")

	err = unstaking(tx.Body, senderState, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, senderState.GetBalanceBigInt().Bytes(), types.MaxAER.Bytes(), "sender.GetBalanceBigInt() cacluation failed")
}

func TestUnstakingError(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.Equal(t, err, nil, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.Equal(t, err, nil, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'u'},
		},
	}
	senderState := &types.State{Balance: types.MaxAER.Bytes()}
	err = unstaking(tx.Body, senderState, scs, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "should be success")
}
