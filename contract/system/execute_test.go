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
			Payload: []byte(`{"Name":"v1stake"}`),
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	emptytx := &types.TxBody{}
	err = ExecuteSystemTx(scs, emptytx, sender, 0)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "Execute system tx failed")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance().Uint64(), uint64(0), "sender.Balance() should be 0 after staking")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	tx.Body.Payload = []byte(`{"Name":"v1voteBP","Args":["16Uiu2HAmBDcLEjBYeEnGU2qDD1KdpEdwDBtN7gqXzNZbHXo8Q841"]}`)
	tx.Body.Amount = big.NewInt(0).Bytes()
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay)
	assert.NoError(t, err, "Execute system tx failed in voting")

	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	tx.Body.Amount = types.StakingMinimum.Bytes()
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay+StakingDelay)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, types.StakingMinimum.Bytes(), sender.Balance().Bytes(),
		"sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(staking.Amount), "check amount of staking")
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
			Payload: buildStakingPayload(false),
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	senderBalance := big.NewInt(0).Add(types.StakingMinimum, types.StakingMinimum)
	sender.AddBalance(senderBalance)

	emptytx := &types.TxBody{}
	err = ExecuteSystemTx(scs, emptytx, sender, 0)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "should error")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.Error(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance(), senderBalance, "sender.Balance() should not chagned after failed unstaking ")

	tx.Body.Payload = buildStakingPayload(true)
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, 0)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	tx.Body.Payload = buildVotingPayload(1)
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay)
	assert.NoError(t, err, "Execute system tx failed in voting")

	tx.Body.Payload = buildStakingPayload(false)
	tx.Body.Amount = senderBalance.Bytes()
	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay+StakingDelay)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance(), senderBalance,
		"sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, VotingDelay+StakingDelay)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Execute system tx failed in unstaking")
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
			Payload: buildStakingPayload(true),
		},
	}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, 0)
	assert.NoError(t, err, "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
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
			Payload: buildStakingPayload(false),
		},
	}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, 0)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, 0)
	assert.EqualError(t, err, types.ErrMustStakeBeforeUnstake.Error(), "Validate system tx failed")

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, 0)
	assert.NoError(t, err, "could not execute system tx")

	tx.Body.Amount = types.StakingMinimum.Bytes()
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, StakingDelay-1)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "Validate system tx failed")
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, StakingDelay)
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
			Payload: buildVotingPayload(0),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, 1)
	assert.EqualError(t, err, types.ErrMustStakeBeforeVote.Error(), "Execute system tx failed")
	tx.Body.Payload = append(tx.Body.Payload, candidates...)

	sender, err := sdb.GetAccountStateV(tx.Body.Account)
	assert.NoError(t, err, "could not get test address state")
	sender.AddBalance(types.StakingMinimum)

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}

	unStakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: buildStakingPayload(false),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	var blockNo uint64
	blockNo = 1
	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "could not execute system tx")

	blockNo += StakingDelay
	err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, blockNo)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "2nd staking tx")

	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	assert.Error(t, err, "empty vote should not allowed")

	tx.Body.Payload = buildVotingPayload(10)
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	assert.NoError(t, err, "fisrt voting validation should success")

	err = ExecuteSystemTx(scs, tx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "fisrt voting execution should success")

	blockNo++
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	assert.Error(t, err, "not enough delay, voting should fail")

	blockNo += VotingDelay
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	assert.NoError(t, err, "after delay, voting should success")

	tx.Body.Payload[1] = '2'
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	t.Log(err.Error())
	assert.NotNil(t, err, "failed to validate system tx for voting")

	tx.Body.Payload = append(tx.Body.Payload, 'i')
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockNo)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "failed to validate system tx for voting")

	blockNo += StakingDelay
	err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, blockNo)
	assert.NoError(t, err, "should execute unstaking system tx")
}
