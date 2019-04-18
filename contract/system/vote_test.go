/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
)

var cdb *state.ChainStateDB
var sdb *state.StateDB

func initTest(t *testing.T) (*state.ContractState, *state.V, *state.V) {
	cdb = state.NewChainStateDB()
	cdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()
	sdb = cdb.OpenNewStateDB(cdb.GetRoot())
	err := cdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"

	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")
	sender, err := sdb.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV([]byte(types.AergoSystem))
	assert.NoError(t, err, "could not get test address state")
	return scs, sender, receiver
}

func getSender(t *testing.T, addr string) *state.V {
	account, err := types.DecodeAddress(addr)
	assert.NoError(t, err, "could not decode test address")
	sender, err := sdb.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	return sender
}

func deinitTest() {
	cdb.Close()
	os.RemoveAll("test")
}

func TestVoteResult(t *testing.T) {
	const testSize = 64
	initTest(t)
	defer deinitTest()
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("testUpdateVoteResult")))
	assert.NoError(t, err, "could not open contract state")
	testResult := map[string]*big.Int{}
	for i := 0; i < testSize; i++ {
		to := fmt.Sprintf("%39d", i) //39:peer id length
		testResult[base58.Encode([]byte(to))] = new(big.Int).SetUint64(uint64(i * i))
	}
	err = InitVoteResult(scs, nil)
	assert.NotNil(t, err, "argument should not be nil")
	err = InitVoteResult(scs, testResult)
	assert.NoError(t, err, "failed to InitVoteResult")

	result, err := getVoteResult(scs, defaultVoteKey, 23)
	assert.NoError(t, err, "could not get vote result")

	oldAmount := new(big.Int).SetUint64((uint64)(math.MaxUint64))
	for i, v := range result.Votes {
		oldi := testSize - (i + 1)
		assert.Falsef(t, new(big.Int).SetBytes(v.Amount).Cmp(oldAmount) > 0,
			"failed to sort result old:%v, %d:%v", oldAmount, i, new(big.Int).SetBytes(v.Amount))
		assert.Equalf(t, uint64(oldi*oldi), new(big.Int).SetBytes(v.Amount).Uint64(), "not match amount value")
		oldAmount = new(big.Int).SetBytes(v.Amount)
	}
}

func TestVoteData(t *testing.T) {
	const testSize = 64
	initTest(t)
	defer deinitTest()
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("testSetGetVoteDate")))
	assert.NoError(t, err, "could not open contract state")

	for i := 0; i < testSize; i++ {
		from := fmt.Sprintf("from%d", i)
		to := fmt.Sprintf("%39d", i)
		vote, err := GetVote(scs, []byte(from), defaultVoteKey)
		assert.NoError(t, err, "failed to getVote")
		assert.Zero(t, vote.Amount, "new amount value is already set")
		assert.Nil(t, vote.Candidate, "new candidates value is already set")

		testVote := &types.Vote{Candidate: []byte(to),
			Amount: new(big.Int).SetUint64(uint64(math.MaxInt64 + i)).Bytes()}

		err = setVote(scs, defaultVoteKey, []byte(from), testVote)
		assert.NoError(t, err, "failed to setVote")

		vote, err = GetVote(scs, []byte(from), defaultVoteKey)
		assert.NoError(t, err, "failed to getVote after set")
		assert.Equal(t, uint64(math.MaxInt64+i), new(big.Int).SetBytes(vote.Amount).Uint64(), "invalid amount")
		assert.Equal(t, []byte(to), vote.Candidate, "invalid candidates")
	}
}

func TestBasicStakingVotingUnstaking(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	sender.AddBalance(types.MaxAER)
	tx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  types.StakingMinimum.Bytes(),
		},
	}

	tx.Body.Payload = buildStakingPayload(true)

	context, err := ValidateSystemTx(tx.Body.Account, tx.Body, sender, scs, 0)
	assert.NoError(t, err, "staking validation")
	_, err = staking(tx.Body, sender, receiver, scs, 0, context)
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be reduced after staking")

	tx.Body.Payload = buildVotingPayload(1)
	context, err = ValidateSystemTx(tx.Body.Account, tx.Body, sender, scs, VotingDelay)
	assert.NoError(t, err, "voting failed")
	_, err = voting(tx.Body, sender, receiver, scs, VotingDelay, context)
	assert.NoError(t, err, "voting failed")

	result, err := getVoteResult(scs, defaultVoteKey, 23)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, context.Call.Args[0].(string), base58.Encode(result.GetVotes()[0].Candidate), "invalid candidate in voting result")
	assert.Equal(t, types.StakingMinimum.Bytes(), result.GetVotes()[0].Amount, "invalid amount in voting result")

	tx.Body.Payload = buildStakingPayload(false)
	_, err = ExecuteSystemTx(scs, tx.Body, sender, receiver, VotingDelay)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "unstaking failed")

	context, err = ValidateSystemTx(tx.Body.Account, tx.Body, sender, scs, VotingDelay+StakingDelay)
	assert.NoError(t, err, "unstaking failed")
	_, err = unstaking(tx.Body, sender, receiver, scs, VotingDelay+StakingDelay, context)
	assert.NoError(t, err, "unstaking failed")

	result2, err := getVoteResult(scs, defaultVoteKey, 23)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result2.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, result.GetVotes()[0].Candidate, result2.GetVotes()[0].Candidate, "invalid candidate in voting result")
	assert.Equal(t, []byte{}, result2.GetVotes()[0].Amount, "invalid candidate in voting result")
}

func buildVotingPayload(count int) []byte {
	var ci types.CallInfo
	ci.Name = types.VoteBP
	for i := 0; i < count; i++ {
		peerID := make([]byte, PeerIDLength)
		peerID[0] = byte(i)
		ci.Args = append(ci.Args, base58.Encode(peerID))
	}
	payload, _ := json.Marshal(ci)
	return payload
}

func buildVotingPayloadEx(count int, name string) []byte {
	var ci types.CallInfo
	ci.Name = name
	switch name {
	case types.VoteBP:
		for i := 0; i < count; i++ {
			_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
			pid, _ := types.IDFromPublicKey(pub)
			ci.Args = append(ci.Args, types.IDB58Encode(pid))
		}
	}
	payload, _ := json.Marshal(ci)
	return payload
}

func buildStakingPayload(isStaking bool) []byte {
	if isStaking {
		return []byte(`{"Name":"v1stake"}`)
	}
	return []byte(`{"Name":"v1unstake"}`)
}
