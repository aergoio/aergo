/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB

func initTest(t *testing.T) {
	sdb = state.NewChainStateDB()
	sdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()

	err := sdb.SetGenesis(genesis)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
}

func deinitTest() {
	sdb.Close()
	os.RemoveAll("test")
}

func TestVoteResult(t *testing.T) {
	const testSize = 64
	initTest(t)
	defer deinitTest()
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("testUpdateVoteResult")))
	assert.NoError(t, err, "could not open contract state")
	testResult := &map[string]uint64{}
	for i := 0; i < testSize; i++ {
		to := fmt.Sprintf("%39d", i) //39:peer id length
		(*testResult)[base58.Encode([]byte(to))] = uint64(i * i)
	}
	err = InitVoteResult(scs, nil)
	assert.NotNil(t, err, "argument should not nil")
	err = InitVoteResult(scs, testResult)
	assert.NoError(t, err, "failed to InitVoteResult")

	const getTestSize = 23
	result, err := GetVoteResult(scs)
	assert.NoError(t, err, "could not get vote result")

	oldAmount := (uint64)(math.MaxUint64)
	for i, v := range result.Votes {
		oldi := testSize - (i + 1)
		assert.Falsef(t, v.Amount > oldAmount, "failed to sort result old:%d, %d:%d", oldAmount, i, v.Amount)
		assert.Equalf(t, uint64(oldi*oldi), v.Amount, "not match amount value")
		oldAmount = v.Amount
	}
}

func TestVoteData(t *testing.T) {
	const testSize = 64
	initTest(t)
	defer deinitTest()
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("testSetGetVoteDate")))
	assert.NoError(t, err, "could not open contract state")

	for i := 0; i < testSize; i++ {
		from := fmt.Sprintf("from%d", i)
		to := fmt.Sprintf("%39d", i)
		vote, err := GetVote(scs, []byte(from))
		assert.NoError(t, err, "failed to getVote")
		assert.Zero(t, vote.Amount, "new amount value is already set")
		assert.Nil(t, vote.Candidate, "new candidates value is already set")

		testVote := &types.Vote{Candidate: []byte(to), Amount: uint64(math.MaxInt64 + i)}

		err = setVote(scs, []byte(from), testVote)
		assert.NoError(t, err, "failed to setVote")

		vote, err = getVote(scs, []byte(from))
		assert.NoError(t, err, "failed to getVote after set")
		assert.Equal(t, uint64(math.MaxInt64+i), vote.Amount, "invalid amount")
		assert.Equal(t, []byte(to), vote.Candidate, "invalid candidates")
	}
}

func TestBasicStakingVotingUnstaking(t *testing.T) {
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
			Amount:  5000,
		},
	}
	senderState := &types.State{Balance: 10000}
	tx.Body.Payload = buildStakingPayload(true)
	err = staking(tx.Body, senderState, scs, 0)
	assert.Equal(t, err, nil, "staking failed")
	assert.Equal(t, senderState.GetBalance(), uint64(5000), "sender balance should be reduced after staking")

	tx.Body.Payload = buildVotingPayload(1)
	err = voting(tx.Body, scs, VotingDelay-1)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "voting failed")

	err = voting(tx.Body, scs, VotingDelay)
	assert.NoError(t, err, "voting failed")

	result, err := GetVoteResult(scs)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, tx.Body.Payload[1:], result.GetVotes()[0].Candidate, "invalid candidate in voting result")
	assert.EqualValues(t, 5000, result.GetVotes()[0].Amount, "invalid amount in voting result")

	tx.Body.Payload = buildStakingPayload(false)
	err = unstaking(tx.Body, senderState, scs, VotingDelay)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "unstaking failed")

	err = unstaking(tx.Body, senderState, scs, VotingDelay+StakingDelay)
	assert.NoError(t, err, "unstaking failed")

	result2, err := GetVoteResult(scs)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result2.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, result.GetVotes()[0].Candidate, result2.GetVotes()[0].Candidate, "invalid candidate in voting result")
	assert.EqualValues(t, 0, result2.GetVotes()[0].Amount, "invalid amount in voting result")
}

func buildVotingPayload(count int) []byte {
	payload := make([]byte, 1+PeerIDLength*count)
	for i := range payload {
		payload[i] = byte(i)
	}
	payload[0] = 'v'
	return payload
}

func buildStakingPayload(isStaking bool) []byte {
	payload := make([]byte, 1)
	if isStaking {
		payload[0] = 's'
	} else {
		payload[0] = 'u'
	}
	return payload
}
