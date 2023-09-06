/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestBasicExecute(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender.ID(),
			Recipient: []byte(types.AergoSystem),
			Amount:    types.StakingMinimum.Bytes(),
			Payload:   []byte(`{"Name":"v1stake"}`),
		},
	}
	sender.AddBalance(types.StakingMinimum)
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	emptytx := &types.TxBody{}
	_, err := ExecuteSystemTx(scs, emptytx, sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "Execute system tx failed")

	events, err := ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance().Uint64(), uint64(0), "sender.Balance() should be 0 after staking")
	assert.Equal(t, events[0].ContractAddress, types.AddressPadding([]byte(types.AergoSystem)), "check event")
	assert.Equal(t, events[0].EventName, types.Opstake.ID(), "check event")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	tx.Body.Payload = []byte(`{"Name":"v1voteBP","Args":["16Uiu2HAmBDcLEjBYeEnGU2qDD1KdpEdwDBtN7gqXzNZbHXo8Q841", "16Uiu2HAmGFzZFifmEhow88XD9RNgDG5dvmqCyNoxg2t7cwochDAj"]}`)
	tx.Body.Amount = big.NewInt(0).Bytes()
	blockInfo.No += VotingDelay
	events, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in voting")
	assert.Equal(t, events[0].ContractAddress, types.AddressPadding([]byte(types.AergoSystem)), "check event")
	assert.Equal(t, events[0].EventName, types.OpvoteBP.ID(), "check event")
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	tx.Body.Amount = types.StakingMinimum.Bytes()
	blockInfo.No += StakingDelay
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, types.StakingMinimum.Bytes(), sender.Balance().Bytes(),
		"sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(staking.Amount), "check amount of staking")
}

func TestBalanceExecute(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender.ID(),
			Recipient: []byte(types.AergoSystem),
			Amount:    types.StakingMinimum.Bytes(),
			Payload:   []byte(`{"Name":"v1stake"}`),
		},
	}
	balance3 := new(big.Int).Mul(types.StakingMinimum, big.NewInt(3))
	balance2 := new(big.Int).Mul(types.StakingMinimum, big.NewInt(2))
	sender.AddBalance(balance3)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	//staking 1
	//balance 3-1=2
	events, err := ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 0 after staking")
	assert.Equal(t, events[0].ContractAddress, types.AddressPadding([]byte(types.AergoSystem)), "check event")
	assert.Equal(t, events[0].EventName, types.Opstake.ID(), "check event")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")
	assert.Equal(t, types.StakingMinimum, receiver.Balance(), "check amount of staking")

	tx.Body.Payload = []byte(`{"Name":"v1voteBP","Args":["16Uiu2HAmBDcLEjBYeEnGU2qDD1KdpEdwDBtN7gqXzNZbHXo8Q841"]}`)
	tx.Body.Amount = big.NewInt(0).Bytes()

	blockInfo.No += VotingDelay
	//voting when 1
	events, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in voting")
	assert.Equal(t, events[0].ContractAddress, types.AddressPadding([]byte(types.AergoSystem)), "check event")
	assert.Equal(t, events[0].EventName, types.OpvoteBP.ID(), "check event")

	voteResult, err := getVoteResult(scs, defaultVoteKey, 1)
	assert.NoError(t, err, "get vote result")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")

	tx.Body.Payload = []byte(`{"Name":"v1stake"}`)
	tx.Body.Amount = balance2.Bytes()

	blockInfo.No += StakingDelay
	//staking 1+2 = 3
	//balance 2-2 = 0
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, big.NewInt(0), sender.Balance(), "sender.Balance() should be 0 after staking")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, balance3, new(big.Int).SetBytes(staking.Amount), "check amount of staking")
	assert.Equal(t, balance3, receiver.Balance(), "check amount of staking")

	//voting still 1
	voteResult, err = getVoteResult(scs, defaultVoteKey, 1)
	assert.NoError(t, err, "get vote result")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")
	internalVoteResult, err := loadVoteResult(scs, defaultVoteKey)
	assert.NoError(t, err, "get vote result")
	var nilBigInt *big.Int
	nilBigInt = nil
	assert.Equal(t, nilBigInt, internalVoteResult.GetTotal(), "defaultVoteKey did not save the total")
	tx.Body.Payload = []byte(`{"Name":"v1unstake"}`)
	tx.Body.Amount = types.StakingMinimum.Bytes()
	blockInfo.No += StakingDelay
	//unstaking 3-1 = 2
	//balance 0+1 = 1
	//voting still 1
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(sender.Balance().Bytes()), "sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, balance2, new(big.Int).SetBytes(staking.Amount), "check amount of staking")
	assert.Equal(t, balance2, receiver.Balance(), "check amount of staking")
	voteResult, err = getVoteResult(scs, defaultVoteKey, 1)
	assert.NoError(t, err, "get vote reulst")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")

	//unstaking 2-3 = -1(fail)
	//balance 1
	//voting 1
	tx.Body.Amount = balance3.Bytes()
	blockInfo.No += StakingDelay
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrExceedAmount, err.Error(), "should return exceed error")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(sender.Balance().Bytes()), "sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, balance2, new(big.Int).SetBytes(staking.Amount), "check amount of staking")
	voteResult, err = getVoteResult(scs, defaultVoteKey, 1)
	assert.NoError(t, err, "get vote reulst")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")

	tx.Body.Amount = balance2.Bytes()
	blockInfo.No += StakingDelay
	//unstaking 2-2 = 0
	//balance 1+2 = 3
	//voting 0
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, balance3, new(big.Int).SetBytes(sender.Balance().Bytes()), "sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(staking.Amount), "check amount of staking")
	voteResult, err = getVoteResult(scs, defaultVoteKey, 1)
	assert.NoError(t, err, "get vote reulst")
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")
}

func TestBasicFailedExecute(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender.ID(),
			Recipient: []byte(types.AergoSystem),
			Amount:    types.StakingMinimum.Bytes(),
			Payload:   buildStakingPayload(false),
		},
	}
	senderBalance := big.NewInt(0).Add(types.StakingMinimum, types.StakingMinimum)
	sender.AddBalance(senderBalance)

	emptytx := &types.TxBody{}
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	_, err := ExecuteSystemTx(scs, emptytx, sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "should error")

	//staking 0+1 = 1
	//balance 2-1 = 1
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "Execute system tx failed in unstaking")
	assert.Equal(t, sender.Balance(), senderBalance, "sender.Balance() should not chagned after failed unstaking")

	tx.Body.Payload = buildStakingPayload(true)
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in staking")
	assert.Equal(t, sender.Balance(), types.StakingMinimum, "sender.Balance() should be 0 after staking")
	staking, err := getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	blockInfo.No += (StakingDelay - 1)
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "check staking delay")

	blockInfo.No += VotingDelay
	tx.Body.Payload = buildVotingPayload(1)
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in voting")
	result, err := getVoteResult(scs, defaultVoteKey, 1)
	assert.Equal(t, types.StakingMinimum, result.Votes[0].GetAmountBigInt(), "check vote result")
	tx.Body.Payload = buildStakingPayload(false)
	tx.Body.Amount = senderBalance.Bytes()
	//staking 1-2 = -1 (fail)
	//balance still 1
	blockInfo.No += StakingDelay
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "should failed with exceed error")
	assert.Equal(t, types.StakingMinimum, sender.Balance(),
		"sender.Balance() should be turn back")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	//staking 1-1 = 0
	//balance 1+1 = 2
	tx.Body.Amount = types.StakingMinimum.Bytes()
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "Execute system tx failed in staking")
	staking, err = getStaking(scs, tx.GetBody().GetAccount())
	assert.Equal(t, senderBalance, sender.Balance(),
		"sender.Balance() should be turn back")
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(staking.Amount), "check amount of staking")

	//staking 0-1 = -1 (fail)
	//balance still 2
	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Execute system tx failed in unstaking")
}

func TestValidateSystemTxForStaking(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender.ID(),
			Recipient: receiver.ID(),
			Amount:    types.StakingMinimum.Bytes(),
			Payload:   buildStakingPayload(true),
		},
	}
	sender.AddBalance(types.StakingMinimum)
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), sender, scs, blockInfo)
	assert.NoError(t, err, "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
}

func TestValidateSystemTxForUnstaking(t *testing.T) {
	scs, sender, receiver := initTest(t)
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
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.EqualError(t, types.ErrMustStakeBeforeUnstake, err.Error(), "Validate system tx failed")
	tx.Body.Amount = new(big.Int).Sub(types.StakingMinimum, new(big.Int).SetUint64(1)).Bytes()
	//_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, 0)
	//assert.EqualError(t, err, types.ErrMustStakeBeforeUnstake.Error(), "Validate system tx failed")

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Amount:  types.StakingMinimum.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	sender.AddBalance(types.StakingMinimum)
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	tx.Body.Amount = types.StakingMinimum.Bytes()
	blockInfo.No += (StakingDelay - 1)
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.EqualError(t, types.ErrLessTimeHasPassed, err.Error(), "Validate system tx failed")
	blockInfo.No++
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.NoError(t, err, "failed to validate system tx for unstaking")
}

func TestValidateSystemTxForVoting(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	const testCandidate = "16Uiu2HAmUJhjwotQqm7eGyZh1ZHrVviQJrdm2roQouD329vxZEkx"
	candidates, err := base58.Decode(testCandidate)
	assert.NoError(t, err, "could not decode candidates")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account: account,
			Payload: buildVotingPayload(0),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.EqualError(t, err, types.ErrMustStakeBeforeVote.Error(), "Execute system tx failed")
	tx.Body.Payload = append(tx.Body.Payload, candidates...)

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
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	blockInfo.No += StakingDelay
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "2nd staking tx")

	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.Error(t, err, "empty vote should not allowed")

	tx.Body.Payload = buildVotingPayload(10)
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.NoError(t, err, "fisrt voting validation should success")

	_, err = ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "fisrt voting execution should success")

	blockInfo.No++
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.Error(t, err, "not enough delay, voting should fail")

	blockInfo.No += VotingDelay
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.NoError(t, err, "after delay, voting should success")

	tx.Body.Payload[1] = '2'
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.NotNil(t, err, "failed to validate system tx for voting")

	tx.Body.Payload = append(tx.Body.Payload, 'i')
	_, err = ValidateSystemTx(tx.Body.Account, tx.GetBody(), nil, scs, blockInfo)
	assert.EqualError(t, types.ErrTxInvalidPayload, err.Error(), "failed to validate system tx for voting")

	blockInfo.No += StakingDelay
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "should execute unstaking system tx")
}

func TestRemainStakingMinimum(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))
	sender.AddBalance(balance3)

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stakingTx.Body.Amount = balance0_5.Bytes()
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrTooSmallAmount.Error(), "could not execute system tx")
	//balance 3-1.5=1.5
	//staking 0+1.5=1.5
	stakingTx.Body.Amount = balance1_5.Bytes()
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	blockInfo.No += StakingDelay
	stakingTx.Body.Amount = balance0_5.Bytes()
	//balance 1.5-0.5=1
	//staking 1.5+1.5=3
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	stakingTx.Body.Amount = balance2.Bytes()
	//balance 1-2=-1 (fail)
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "check error")

	stakingTx.Body.Amount = balance1.Bytes()
	//time fail
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "check error")

	unStakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance0_5.Bytes(),
			Payload: buildStakingPayload(false),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	blockInfo.No += (StakingDelay - 1)
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "check error")

	blockInfo.No += StakingDelay
	//balance 1+0.5 =1.5
	//staking 2-0.5 =1.5
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	staked, err := getStaking(scs, sender.ID())
	assert.NoError(t, err, "could not get staking")
	assert.Equal(t, balance1_5, sender.Balance(), "could not get staking")
	assert.Equal(t, balance1_5, staked.GetAmountBigInt(), "could not get staking")

	blockInfo.No += StakingDelay
	//balance 1.5+0.5 =2
	//staking 1.5-0.5 =1
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	staked, err = getStaking(scs, sender.ID())
	assert.NoError(t, err, "could not get staking")
	assert.Equal(t, balance2, sender.Balance(), "could not get staking")
	assert.Equal(t, balance1, staked.GetAmountBigInt(), "could not get staking")

	blockInfo.No += StakingDelay
	//staking 1-0.5 =0.5 (fail)
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrTooSmallAmount.Error(), "staked aergo remain 0.5")
	staked, err = getStaking(scs, sender.ID())
	assert.NoError(t, err, "could not get staking")
	assert.Equal(t, balance2, sender.Balance(), "could not get staking")
	assert.Equal(t, balance1, staked.GetAmountBigInt(), "could not get staking")

	blockInfo.No += StakingDelay
	unStakingTx.Body.Amount = balance1.Bytes()
	//balance 2+1 =3
	//staking 1-1 =0
	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	staked, err = getStaking(scs, sender.ID())
	assert.NoError(t, err, "could not get staking")
	assert.Equal(t, balance3, sender.Balance(), "could not get staking")
	assert.Equal(t, big.NewInt(0), staked.GetAmountBigInt(), "could not get staking")

	_, err = ExecuteSystemTx(scs, unStakingTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrMustStakeBeforeUnstake.Error(), "check error")
}

func TestProposalExecute(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	//balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	//balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))

	sender.AddBalance(balance3)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	blockInfo.No++
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)

	votingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	event, err := ExecuteSystemTx(scs, votingTx.GetBody(), sender, receiver, blockInfo)
	assert.Equal(t, "voteDAO", event[0].EventName, "event name")
	assert.Equal(t, "[\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"BPCOUNT\", {\"_bignum\":\"13\"}]", event[0].JsonArgs, "event args")
	assert.NoError(t, err, "failed in voting proposal")

	vinfo, err := GetVotes(scs, sender.ID())
	assert.NoError(t, err, "failed in get vote")
	t.Log(vinfo)
	assert.Equal(t, "13", string(vinfo[0].Candidates[0]), "check vote")

	v, err := getVote(scs, bpCount.Key(), sender.ID())
	assert.NoError(t, err, "failed in get vote")
	assert.Equal(t, "[\"13\"]", string(v.Candidate), "check vote candidates")
	assert.Equal(t, balance1, new(big.Int).SetBytes(v.Amount), "check vote amount")

	voteResult, err := getVoteResult(scs, bpCount.Key(), 1)
	assert.NoError(t, err, "get vote result")
	assert.Equal(t, types.StakingMinimum, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")

	blockInfo.No += StakingDelay
	unstakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(false),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, unstakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, new(big.Int).Sub(balance3, types.ProposalPrice), sender.Balance(), "sender.Balance() should be 2 after unstaking")

	voteResult, err = getVoteResult(scs, GenProposalKey(bpCount.ID()), 1)
	assert.NoError(t, err, "get vote result")
	assert.Equal(t, big.NewInt(0), new(big.Int).SetBytes(voteResult.Votes[0].Amount), "check result amount")
	assert.Equal(t, 1, len(voteResult.Votes), "check result length")
}

func TestProposalExecuteFail1(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	//balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	//balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))
	sender.AddBalance(balance3)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}

	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	invalidaVersionTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["numbp", "non","13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, invalidaVersionTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "the proposal is not created (numbp, non)")

	//deprecated
	tooEarlyTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["numbp", "13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, tooEarlyTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "the voting begins at 1")

	blockInfo.No += 10
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)
	tooManyCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13","23","17"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, tooManyCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "too many candidates arguments (max : 1)")

	invalidCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "ab"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, invalidCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "include invalid count")

	blockInfo.No += VotingDelay
	tooLateTx := tooEarlyTx
	_, err = ExecuteSystemTx(scs, tooLateTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "the voting was already done at 10")
}

func TestProposalExecuteFail2(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	//balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	//balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))

	sender.AddBalance(balance3)
	sender2 := getSender(t, "AmNqJN2P1MA2Uc6X5byA4mDg2iuo95ANAyWCmd3LkZe4GhJkSyr4")
	sender2.AddBalance(balance3)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	blockInfo.No++
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)
	validCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["BPCOUNT", "1"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "valid")
	internalVoteResult, err := loadVoteResult(scs, GenProposalKey("BPCOUNT"))
	assert.Equal(t, balance1, internalVoteResult.GetTotal(), "check result total")

}

func TestProposalExecute2(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))
	balance5 := new(big.Int).Mul(balance1, big.NewInt(5))

	sender2 := getSender(t, "AmNqJN2P1MA2Uc6X5byA4mDg2iuo95ANAyWCmd3LkZe4GhJkSyr4")
	sender3 := getSender(t, "AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7")
	sender.AddBalance(balance3)
	sender2.AddBalance(balance3)
	sender3.AddBalance(balance3)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance2.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance1, sender.Balance(), "sender.Balance() should be 1 after staking")

	stakingTx.Body.Account = sender2.ID()
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance1, sender2.Balance(), "sender.Balance() should be 1 after staking")

	stakingTx.Body.Account = sender3.ID()
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance1, sender3.Balance(), "sender.Balance() should be 1 after staking")

	blockInfo.No++
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)

	votingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["BPCOUNT", "23"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "failed in voting proposal")
	votingTx.Body.Account = sender2.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["BPCOUNT", "13"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	votingTx.Body.Account = sender3.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["BPCOUNT", "13"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	voteResult, err := getVoteResult(scs, GenProposalKey("BPCOUNT"), 3)
	assert.NoError(t, err, "get vote result")
	assert.Equal(t, new(big.Int).Mul(balance2, big.NewInt(2)), new(big.Int).SetBytes(voteResult.Votes[0].Amount), "")
	assert.Equal(t, "13", string(voteResult.Votes[0].Candidate), "1st place")
	assert.Equal(t, balance2, new(big.Int).SetBytes(voteResult.Votes[1].Amount), "")
	assert.Equal(t, "23", string(voteResult.Votes[1].Candidate), "2nd place")
	internalVoteResult, err := loadVoteResult(scs, GenProposalKey(bpCount.ID()))
	assert.Equal(t, new(big.Int).Mul(balance2, big.NewInt(3)), internalVoteResult.GetTotal(), "check result total")

	votingTx = &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["STAKINGMIN", "` + balance0_5.String() + `"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "failed in voting proposal")
	votingTx.Body.Account = sender2.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["STAKINGMIN", "` + balance0_5.String() + `"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	votingTx.Body.Account = sender3.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["STAKINGMIN", "10000"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	votingTx = &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["GASPRICE", "` + balance0_5.String() + `"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "failed in voting proposal")
	votingTx.Body.Account = sender2.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["GASPRICE", "` + balance0_5.String() + `"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	votingTx.Body.Account = sender3.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["GASPRICE", "1004"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	gasPrice := GetGasPrice()
	assert.Equal(t, balance0_5, gasPrice, "result of gas price voting")

	blockInfo.No += StakingDelay
	unstakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(false),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, unstakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, new(big.Int).Sub(balance2, types.ProposalPrice), sender.Balance(), "sender.Balance() should be 2 after unstaking")

	voteResult, err = getVoteResult(scs, GenProposalKey(bpCount.ID()), 3)
	assert.NoError(t, err, "get vote result")
	internalVoteResult, err = loadVoteResult(scs, GenProposalKey(bpCount.ID()))
	assert.Equal(t, balance5, internalVoteResult.GetTotal(), "check result total")

	assert.Equal(t, new(big.Int).Mul(balance2, big.NewInt(2)), new(big.Int).SetBytes(voteResult.Votes[0].Amount), "check result amount")
	assert.Equal(t, "13", string(voteResult.Votes[0].Candidate), "1st place")
	assert.Equal(t, balance1, new(big.Int).SetBytes(voteResult.Votes[1].Amount), "check result amount")
	assert.Equal(t, "23", string(voteResult.Votes[1].Candidate), "2nd place")

	scs = commitNextBlock(t, scs)
	blockInfo.No += StakingDelay

	unstakingTx.Body.Amount = balance0_5.Bytes()
	_, err = ExecuteSystemTx(scs, unstakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	votingTx.Body.Account = sender2.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["NAMEPRICE", "1004"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	votingTx.Body.Account = sender3.ID()
	votingTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["NAMEPRICE", "1004"]}`)
	_, err = ExecuteSystemTx(scs, votingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")

	voteResult, err = getVoteResult(scs, GenProposalKey(namePrice.ID()), 3)
	assert.NoError(t, err, "get vote result")
	internalVoteResult, err = loadVoteResult(scs, GenProposalKey(namePrice.ID()))
	assert.Equal(t, new(big.Int).Mul(balance2, big.NewInt(2)), internalVoteResult.GetTotal(), "check result total")
	assert.Equal(t, "1004", string(voteResult.Votes[0].Candidate), "1st place")
	currentNamePrice := GetNamePrice()
	assert.Equal(t, "1004", currentNamePrice.String(), "current name price")

	/*
		blockInfo += StakingDelay
		//voting result was freeze
		_, err = ExecuteSystemTx(scs, unstakingTx.GetBody(), sender, receiver, blockInfo)
		assert.NoError(t, err, "could not execute system tx")
		assert.Equal(t, new(big.Int).Sub(balance3, types.ProposalPrice), sender.Balance(), "sender.Balance() should be 2 after unstaking")

		voteResult, err = getVoteResult(scs, GenProposalKey("numbp"), 3)
		assert.NoError(t, err, "get vote result")
		internalVoteResult, err = loadVoteResult(scs, GenProposalKey("numbp"))
		assert.Equal(t, balance5, internalVoteResult.GetTotal(), "check result total")
		assert.Equal(t, balance5, new(big.Int).SetBytes(voteResult.Votes[0].Amount), "check result amount")
		assert.Equal(t, "13", string(voteResult.Votes[0].Candidate), "1st place")
		assert.Equal(t, balance3, new(big.Int).SetBytes(voteResult.Votes[1].Amount), "check result amount")
		assert.Equal(t, "23", string(voteResult.Votes[1].Candidate), "2nd place")
		assert.Equal(t, balance2, new(big.Int).SetBytes(voteResult.Votes[2].Amount), "check result amount")
		assert.Equal(t, "27", string(voteResult.Votes[2].Candidate), "1st place")
	*/
}
