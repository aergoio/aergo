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

	err = unstaking(tx.Body, senderState, scs, 1)
	assert.Equal(t, err, ErrLessTimeHasPassed, "should be return ErrLessTimeHasPassed")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	err = unstaking(tx.Body, senderState, scs, 5)
	assert.Equal(t, err, nil, "should be success")
	assert.Equal(t, senderState.GetBalance(), uint64(1000), "sender balance cacluation failed")
}

func TestStaking1000Unstaking9000(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	if err != nil {
		t.Error("could not open contract state")
	}

	account, err := types.DecodeAddress(testSender)
	if err != nil {
		t.Fatalf("could not decode test Address :%s", err.Error())
	}

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  1000,
		},
	}
	senderState := &types.State{Balance: 1000}
	err = staking(tx.Body, senderState, scs, 0)
	if err != nil {
		t.Errorf("failed to staking: %s", err.Error())
	}
	if senderState.GetBalance() != 0 {
		t.Errorf("sender balance calculation error : %d", senderState.GetBalance())
	}
	tx.Body.Amount = 9000
	err = unstaking(tx.Body, senderState, scs, 1)
	if err != ErrLessTimeHasPassed {
		t.Errorf("should be return ErrLessTimeHasPassed but %s", err.Error())
	}
	if senderState.GetBalance() != 0 {
		t.Errorf("sender balance calculation error : %d", senderState.GetBalance())
	}
	err = unstaking(tx.Body, senderState, scs, 5)
	if err != nil {
		t.Errorf("should be success but %s", err.Error())
	}
	if senderState.GetBalance() != 1000 {
		t.Errorf("sender balance calculation error : %d", senderState.GetBalance())
	}
}
