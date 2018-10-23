/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
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
			Amount:  1000,
			Payload: []byte{'s'},
		},
	}
	senderState := &types.State{Balance: 1000}

	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	tx.Body.Payload = []byte{'u'}
	err = unstaking(tx.Body, senderState, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, senderState.GetBalance(), uint64(1000), "sender balance cacluation failed")
}

func TestStaking1000Unstaking9000(t *testing.T) {
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
			Amount:  1000,
			Payload: []byte{'s'},
		},
	}
	senderState := &types.State{Balance: 1000}
	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	tx.Body.Amount = 9000
	tx.Body.Payload = []byte{'u'}
	err = unstaking(tx.Body, senderState, scs, StakingDelay-1)
	assert.Equal(t, err, types.ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, StakingDelay)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, senderState.GetBalance(), uint64(1000), "sender balance cacluation failed")
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
			Amount:  1000,
			Payload: []byte{'u'},
		},
	}
	senderState := &types.State{Balance: 1000}
	err = unstaking(tx.Body, senderState, scs, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "should be success")
}
