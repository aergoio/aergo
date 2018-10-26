/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestBasicExecute(t *testing.T) {
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

	emptytx := &types.TxBody{}
	err = ExecuteSystemTx(emptytx, senderState, scs, 0)
	assert.EqualError(t, types.ErrTxFormatInvalid, err.Error(), "Execute system tx failed")

	err = ExecuteSystemTx(tx.GetBody(), senderState, scs, 0)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, senderState.GetBalance(), uint64(0), "sender balance should be 0 after staking")

	tx.Body.Payload = []byte{'v'}
	err = ExecuteSystemTx(tx.GetBody(), senderState, scs, VotingDelay)
	assert.NoError(t, err, "Execute system tx failed in voting")

	tx.Body.Payload = []byte{'u'}
	err = ExecuteSystemTx(tx.GetBody(), senderState, scs, VotingDelay+StakingDelay)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, senderState.GetBalance(), uint64(1000), "sender balance should be 0 after staking")
}

func TestValidateSystemTxForStaking(t *testing.T) {
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
			Amount:  types.StakingMinimum,
			Payload: []byte{'s'},
		},
	}
	err = ValidateSystemTx(tx.GetBody(), scs, 0)
	assert.NoError(t, err, "Execute system tx failed")
	tx.Body.Amount = types.StakingMinimum - 1
	err = ValidateSystemTx(tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrTooSmallAmount, err.Error(), "Execute system tx failed")
}

func TestValidateSystemTxForUnstaking(t *testing.T) {
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
			Amount:  types.StakingMinimum,
			Payload: []byte{'u'},
		},
	}
	err = ValidateSystemTx(tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Execute system tx failed")
	tx.Body.Amount = types.StakingMinimum - 1
	err = ValidateSystemTx(tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrTooSmallAmount, err.Error(), "Execute system tx failed")

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  1000,
			Payload: []byte{'s'},
			Type:    types.TxType_GOVERNANCE,
		},
	}
	senderState := &types.State{Balance: 1000}
	err = ExecuteSystemTx(stakingTx.GetBody(), senderState, scs, 0)
	assert.NoError(t, err, "could not execute system tx")

	tx.Body.Amount = types.StakingMinimum
	err = ValidateSystemTx(tx.GetBody(), scs, StakingDelay-1)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "Execute system tx failed")
	err = ValidateSystemTx(tx.GetBody(), scs, StakingDelay)
	assert.NoError(t, err, "failed to validate system tx for unstaking")
}

func TestValidateSystemTxForVoting(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	const testCandidate = "16Uiu2HAmUJhjwotQqm7eGyZh1ZHrVviQJrdm2roQouD329vxZEkx"
	candidates, err := base58.Decode(testCandidate)
	assert.NoError(t, err, "could not decode candidates")
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Payload: []byte{'v'},
			Type:    types.TxType_GOVERNANCE,
		},
	}
	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay)
	assert.EqualError(t, types.ErrMustStakeBeforeVote, err.Error(), "Execute system tx failed")

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  1000,
			Payload: []byte{'s'},
			Type:    types.TxType_GOVERNANCE,
		},
	}
	senderState := &types.State{Balance: 1000}
	err = ExecuteSystemTx(stakingTx.GetBody(), senderState, scs, 0)
	assert.NoError(t, err, "could not execute system tx")

	tx.Body.Payload = append(tx.Body.Payload, candidates...)

	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay-1)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "failed to validate system tx")

	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay)
	assert.NoError(t, err, "failed to validate system tx for voting")

	tx.Body.Payload[0] = 'v'
	err = ExecuteSystemTx(stakingTx.GetBody(), senderState, scs, VotingDelay)
	assert.NoError(t, err, "could not execute system tx")

	err = ExecuteSystemTx(stakingTx.GetBody(), senderState, scs, VotingDelay+StakingDelay)
	assert.NoError(t, err, "could not execute system tx")

	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay+StakingDelay+1)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "failed to validate system tx")

	tx.Body.Payload[1] = '2'
	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay)
	t.Log(err.Error())
	assert.NotNil(t, err, "failed to validate system tx for voting")

	tx.Body.Payload = append(tx.Body.Payload, 'i')
	err = ValidateSystemTx(tx.GetBody(), scs, VotingDelay)
	assert.EqualError(t, types.ErrTxFormatInvalid, err.Error(), "failed to validate system tx for voting")

}
