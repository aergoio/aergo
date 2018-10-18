/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/aergoio/aergo/state"
	"github.com/stretchr/testify/assert"

	"github.com/aergoio/aergo/types"
)

var sdb *state.ChainStateDB

func initTest(t *testing.T) {
	sdb = state.NewChainStateDB()
	sdb.Init("test", nil)
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

	for i := 0; i < testSize; i++ {
		to := fmt.Sprintf("%39d", i) //39:peer id length
		err := updateVoteResult(scs, []byte(to), (uint64)(i*i), true)
		assert.NoError(t, err, "failed to updateVoteResult")
	}
	const getTestSize = 23
	result, err := getVoteResult(scs, getTestSize)
	assert.NoError(t, err, "could not get vote result")

	oldAmount := (uint64)(math.MaxUint64)
	for i, v := range result.Votes {
		oldi := testSize - (i + 1)
		assert.Falsef(t, v.Amount > oldAmount, "failed to sort result old:%d, %d:%d", oldAmount, i, v.Amount)
		assert.Equalf(t, uint64(oldi*oldi), v.Amount, "not match amount value", oldAmount, i, v.Amount)
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
		amount, updateBlockNo, candidates, err := getVote(scs, []byte(from))
		assert.NoError(t, err, "failed to getVote")
		assert.Zero(t, amount, "new amount value is already set")
		assert.Zero(t, updateBlockNo, "new updateBlockNo value is already set")
		assert.Nil(t, candidates, "new candidates value is already set")

		err = setVote(scs, []byte(from), []byte(to), (uint64)(math.MaxInt64+i), (uint64)(i))
		assert.NoError(t, err, "failed to setVote")

		amount, updateBlockNo, candidates, err = getVote(scs, []byte(from))
		assert.NoError(t, err, "failed to getVote after set")
		assert.Equal(t, uint64(math.MaxInt64+i), amount, "invalid amount")
		assert.Equal(t, uint64(i), updateBlockNo, "invalid block number")
		assert.Equal(t, []byte(to), candidates, "invalid candidates")
	}
}
