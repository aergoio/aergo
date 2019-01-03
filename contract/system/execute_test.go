/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestBasicExecute(t *testing.T) {
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
			Payload: []byte{'s'},
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	emptytx := &types.TxBody{}
	err = ExecuteSystemTx(scs, emptytx, sender, 0)
	assert.EqualError(t, types.ErrTxFormatInvalid, err.Error(), "Execute system tx failed")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance().Uint64(), uint64(0), "sender.Balance() should be 0 after staking")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	tx.Body.Payload = []byte{'v'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay)
	assert.NoError(t, err, "Execute system tx failed in voting")

	tx.Body.Payload = []byte{'u'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay+StakingDelay)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance().Bytes(), types.StakingMinimum.Bytes(),
		"sender.Balance() should be turn back")
}

func TestBasicFailedExecute(t *testing.T) {
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
			Payload: []byte{'u'},
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	emptytx := &types.TxBody{}
	err = ExecuteSystemTx(scs, emptytx, sender, 0)
	assert.EqualError(t, types.ErrTxFormatInvalid, err.Error(), "Execute system tx failed")

	tx.Body.Payload = []byte{'u'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.Error(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should not chagned after failed unstaking ")

	tx.Body.Payload = []byte{'s'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance().Uint64(), uint64(0), "sender.Balance() should be 0 after staking")

	tx.Body.Payload = []byte{'v'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay)
	assert.NoError(t, err, "Execute system tx failed in voting")

	tx.Body.Payload = []byte{'u'}
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay+StakingDelay)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance().Bytes(), types.StakingMinimum.Bytes(),
		"sender.Balance() should be turn back")
}

func TestValidateSystemTxForStaking(t *testing.T) {
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
			Payload: []byte{'s'},
		},
	}
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, 0)
	assert.NoError(t, err, "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrTooSmallAmount, err.Error(), "Validate system tx failed")
}

func TestValidateSystemTxForUnstaking(t *testing.T) {
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
			Payload: []byte{'u'},
		},
	}
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, 0)
	assert.EqualError(t, types.ErrTooSmallAmount, err.Error(), "Validate system tx failed")

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'s'},
			Type:    types.TxType_GOVERNANCE,
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, 0)
	assert.NoError(t, err, "could not execute system tx")

	tx.Body.Amount = types.StakingMinimum.Bytes()
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, StakingDelay-1)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "Validate system tx failed")
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, StakingDelay)
	assert.NoError(t, err, "failed to validate system tx for unstaking")
}

func TestValidateSystemTxForVoting(t *testing.T) {
	initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	const testCandidate = "16Uiu2HAmUJhjwotQqm7eGyZh1ZHrVviQJrdm2roQouD329vxZEkx"
	candidates, err := base58.Decode(testCandidate)
	assert.NoError(t, err, "could not decode candidates")
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
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
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, 1)
	assert.EqualError(t, types.ErrMustStakeBeforeVote, err.Error(), "Execute system tx failed")
	tx.Body.Payload = append(tx.Body.Payload, candidates...)

	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'s'},
			Type:    types.TxType_GOVERNANCE,
		},
	}

	unStakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: []byte{'u'},
			Type:    types.TxType_GOVERNANCE,
		},
	}
	var blockNo uint64
	blockNo = 1
	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "could not execute system tx")

	blockNo += StakingDelay
	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "could not execute system tx")

	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, blockNo)
	assert.NoError(t, err, "fisrt voting validation should success")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "fisrt voting execution should success")

	blockNo++
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, blockNo)
	assert.Error(t, err, "not enough delay, voting should fail")

	blockNo += VotingDelay
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, blockNo)
	assert.NoError(t, err, "after delay, voting should success")

	tx.Body.Payload[1] = '2'
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, blockNo)
	t.Log(err.Error())
	assert.NotNil(t, err, "failed to validate system tx for voting")

	tx.Body.Payload = append(tx.Body.Payload, 'i')
	err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), scs, blockNo)
	assert.EqualError(t, types.ErrTxFormatInvalid, err.Error(), "failed to validate system tx for voting")

	blockNo += StakingDelay
	err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "should execute unstaking system tx")
}
