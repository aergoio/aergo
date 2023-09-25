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
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
)

var cdb *state.ChainStateDB
var bs *state.BlockState

func initTest(t *testing.T) (*state.ContractState, *state.V, *state.V) {
	cdb = state.NewChainStateDB()
	cdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()
	err := cdb.SetGenesis(genesis, nil)

	bs = cdb.NewBlockState(cdb.GetRoot())
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	// Need to pass the
	InitGovernance("dpos")
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"

	scs, err := bs.OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	InitSystemParams(scs, 3)

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")
	sender, err := bs.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := bs.GetAccountStateV([]byte(types.AergoSystem))
	assert.NoError(t, err, "could not get test address state")
	return scs, sender, receiver
}

func getSender(t *testing.T, addr string) *state.V {
	account, err := types.DecodeAddress(addr)
	assert.NoError(t, err, "could not decode test address")
	sender, err := bs.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	return sender
}

func commitNextBlock(t *testing.T, scs *state.ContractState) *state.ContractState {
	bs.StageContractState(scs)
	bs.Update()
	bs.Commit()
	cdb.UpdateRoot(bs)
	systemContractID := types.ToAccountID([]byte(types.AergoSystem))
	systemContract, err := bs.GetAccountState(systemContractID)
	assert.NoError(t, err, "could not get account state")
	ret, err := bs.OpenContractState(systemContractID, systemContract)
	assert.NoError(t, err, "could not open contract state")
	return ret
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

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stake, err := newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "staking validation")
	event, err := stake.run()
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be reduced after staking")
	assert.Equal(t, event.EventName, "stake", "event name")
	assert.Equal(t, event.JsonArgs, "{\"who\":\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"amount\":\"10000000000000000000000\"}", "event args")

	tx.Body.Payload = buildVotingPayload(1)
	blockInfo.No += VotingDelay
	voting, err := newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "voting failed")
	event, err = voting.run()
	assert.NoError(t, err, "voting failed")
	assert.Equal(t, event.EventName, "voteBP", "event name")
	assert.Equal(t, event.JsonArgs, "{\"who\":\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"vote\":[\"111111111111111111111111111111111111111\"]}")

	result, err := getVoteResult(scs, defaultVoteKey, 23)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, voting.arg(0), base58.Encode(result.GetVotes()[0].Candidate), "invalid candidate in voting result")
	assert.Equal(t, types.StakingMinimum.Bytes(), result.GetVotes()[0].Amount, "invalid amount in voting result")

	tx.Body.Payload = buildStakingPayload(false)
	_, err = ExecuteSystemTx(scs, tx.Body, sender, receiver, blockInfo)
	assert.EqualError(t, err, types.ErrLessTimeHasPassed.Error(), "unstaking failed")

	blockInfo.No += StakingDelay
	unstake, err := newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "unstaking failed")
	event, err = unstake.run()
	assert.NoError(t, err, "unstaking failed")
	assert.Equal(t, event.EventName, "unstake", "event name")
	assert.Equal(t, event.JsonArgs, "{\"who\":\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", \"amount\":\"10000000000000000000000\"}", "event args")

	result2, err := getVoteResult(scs, defaultVoteKey, 23)
	assert.NoError(t, err, "voting failed")
	assert.EqualValues(t, len(result2.GetVotes()), 1, "invalid voting result")
	assert.Equal(t, result.GetVotes()[0].Candidate, result2.GetVotes()[0].Candidate, "invalid candidate in voting result")
	assert.Equal(t, []byte{}, result2.GetVotes()[0].Amount, "invalid candidate in voting result")

	blockInfo.No += StakingDelay
	blockInfo.ForkVersion = 2
	tx.Body.Payload = buildStakingPayload(true)
	stake, err = newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "staking validation")
	event, err = stake.run()
	assert.NoError(t, err, "staking failed")
	assert.Equal(t, sender.Balance().Bytes(), new(big.Int).Sub(types.MaxAER, types.StakingMinimum).Bytes(),
		"sender.Balance() should be reduced after staking")
	assert.Equal(t, event.EventName, "stake", "event name")
	assert.Equal(t, event.JsonArgs, "[\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", {\"_bignum\":\"10000000000000000000000\"}]", "event args")

	tx.Body.Payload = buildVotingPayload(30)
	blockInfo.No += VotingDelay
	voting, err = newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	assert.NoError(t, err, "voting failed")
	event, err = voting.run()
	assert.NoError(t, err, "voting failed")
	assert.Equal(t, event.EventName, "voteBP", "event name")
	assert.Equal(t, `["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4", ["111111111111111111111111111111111111111","esSRNKFHpMYCTUPeaQ3coZDgiERzi8R6g6UNFHhEhVwD5jvYV81M","2JjsqjdVadi5PuwnJ9o6Ec7SNRTrzRFqCMBwjVaPUPzsRAUr5yF1h","2xcKG6wjsT4cbNRAwjC8rQff48hHz8PFJ2HR6js5i6VodFDmdTN23","3cUkgUFzAGR9nptZbJbBUDDsjqviyqWfPhNtTz9mwnzjqKxhAwV2P","4GMC6qaET5mgzHMxEszE61n6RZA9yYe5VNUMqESUBVVg3QhciRc2j","4vDdXCtUju8EBjqLtTPGhpLK7GPayFmVb3ZqCUjARBzcFVSYFuj35","5a64waCj2iUmPCJjY2nKKctXnyd1xxtugifJZj1retVYTaBToPr3R","6DxWMwWyKXqJaen8BcBMwRSkUgrSxg2KnPkmvyJYtazUfevPLsy3m","6spwnJqDcMBqn7FWqBaQZDzyAQ5sxP9jt4rFJDbF8HVQsjfJtN647","7XhPCg9TuAYNyZiuUkyTB2ZBr7KJx6H9yjwifTswMyzM5pQERrD4T","8BZpd3TiBytvB2CJ8LNVnq7QXpYjwoQa5R3C2iAdbgVHHu99yLL4o","8qSG3QmxUoFTNUfgmumYQdfdDXnAwWXzB68fPxTKqNzDVyt5WpT59","9VJhTn6CmcbzZw95RVAb2SDquF1bwDfQGmE8mCk255V9i4d14Ja5V","A9B8t9QT4RxXmPcU54ZdeEn4axF2vvnpNSKc8T2iJmz5v9Mvbnh5q","Ao3aJWihMFK4xr5ridxgG3LHGfUTvdvEU7R5VhKQYUV28E6r9Gp6B","BSv1it2we4fcAJZFNDMisqtVxNhtvM3eZnWYrwc6nAyxLJqmgkw6X","C6nT9FMBvt29Mm2e1nkmVeSie5wKv4B4fTc2EBto1sUtYPahEF46s","CketZcfSDhNgZDW2fN9p7SzwKoAkumJUm8hVbSBVFZypkUKcmjB7D","DQXKyyygWWjDkfyRJwYrjFZA1WQBuURtronxxgUBVGUkxZ4YKDJ7Z","E4PmQMHvoL5kx8SoxWwuM47NhDdcuBZJxUtSKvksixyhAdoTrhR7u","EiGCpicB69SJ9avCc6LwxrfbNvs3ttgj49yuhB3ZxfUdNiYPQBY8F","FN8eF5vRNxnqM3PbFfjzafDp4e6Utbp99q5P4RLGCMyZaoHJwff8b","G215fTEffn9NYVryuF93CTn2kMKutJwZFWArRfcxS4UVnt2EV9n8w","GfsX5pYuxbVujxLNYpY5pGLFS4ZLt24yMBGKnuuefkyRzxmA2du9H","HKjxWBsAFQrSwQomCPw8S4tU7mnmsjCPSrMoAACLuTUND3W5a829d","HycPvZBQYECz8sH9qyLB3sSgoV2CsSKoYXTGXQV399yJR8F17c99y","JdUqLvVeq3ZXLKkYVYjDffzuVCFds9TDeCYjtemjNrUEdCyvf6GAK","KHMGmHou7rv4XnDw988GHUZ8AuV4rradjseDFu4RcYyAqHirCaPAf","KwDiBf89QgGbjEhKnhXJuH7LrciVrZi3qYjgd9M7rFU73NTmk4WB1"]]`,
		event.JsonArgs, "event args")

	blockInfo.No += StakingDelay
	tx.Body.Payload = buildStakingPayload(false)
	unstake, err = newSysCmd(tx.Body.Account, tx.Body, sender, receiver, scs, blockInfo)
	event, err = unstake.run()
	assert.NoError(t, err, "unstaking failed")
	assert.Equal(t, event.EventName, "unstake", "event name")
	assert.Equal(t, event.JsonArgs, "[\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\", {\"_bignum\":\"10000000000000000000000\"}]", "event args")
}

func buildVotingPayload(count int) []byte {
	var ci types.CallInfo
	ci.Name = types.OpvoteBP.Cmd()
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
	switch types.GetOpSysTx(name) {
	case types.OpvoteBP:
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

func TestVotingCatalog(t *testing.T) {
	cat := GetVotingCatalog()
	assert.Equal(t, 5, len(cat))
	for _, issue := range cat {
		fmt.Println(issue.ID())
	}
}
