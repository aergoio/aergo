/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aergoio/aergo/types"
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
		},
	}
	senderState := &types.State{Balance: 1000}

	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, stakingDelay-1)
	assert.Equal(t, err, ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, stakingDelay)
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
		},
	}
	senderState := &types.State{Balance: 1000}
	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	tx.Body.Amount = 9000
	err = unstaking(tx.Body, senderState, scs, stakingDelay-1)
	assert.Equal(t, err, ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, stakingDelay)
	assert.NoError(t, err, "should be success")
	assert.Equal(t, senderState.GetBalance(), uint64(1000), "sender balance cacluation failed")
}
