/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/aergoio/aergo/state"

	"github.com/aergoio/aergo/types"
)

var sdb *state.ChainStateDB

func initTest(t *testing.T) {
	sdb = state.NewChainStateDB()
	sdb.Init("test")
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
	scs, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte("testUpdateVoteResult")))
	if err != nil {
		t.Error("could not open contract state")
	}
	for i := 0; i < testSize; i++ {
		to := fmt.Sprintf("%39d", i) //39:peer id length
		err := updateVoteResult(scs, []byte(to), (uint64)(i*i), true)
		if err != nil {
			t.Errorf("failed to updateVoteResult: %s", err.Error())
		}
	}
	const getTestSize = 23
	result, err := getVoteResult(scs, getTestSize)
	if err != nil {
		t.Errorf("could not get vote result : %s", err.Error())
	}
	oldAmount := (uint64)(math.MaxUint64)
	for i, v := range result.Votes {
		oldi := testSize - (i + 1)
		if v.Amount > oldAmount {
			t.Errorf("failed to sort result old:%d, %d:%d", oldAmount, i, v.Amount)
		}
		if v.Amount != (uint64)(oldi*oldi) {
			t.Errorf("not match amount value old:%d, %d:%d", oldi*oldi, i, v.Amount)
		}
		oldAmount = v.Amount
	}
}

func TestVoteData(t *testing.T) {
	const testSize = 64
	initTest(t)
	defer deinitTest()
	scs, err := sdb.OpenContractStateAccount(types.ToAccountID([]byte("testSetGetVoteDate")))
	if err != nil {
		t.Error("could not open contract state")
	}
	for i := 0; i < testSize; i++ {
		from := fmt.Sprintf("from%d", i)
		to := fmt.Sprintf("%39d", i)
		amount, updateBlockNo, candidates, err := getVote(scs, []byte(from))
		if err != nil {
			t.Errorf("failed to getVote : %s", err.Error())
		}
		if amount != 0 {
			t.Errorf("new amount value is already set : %d", amount)
		}
		if updateBlockNo != 0 {
			t.Errorf("new updateBlockNo value is already set : %d", updateBlockNo)
		}
		if candidates != nil {
			t.Errorf("new candidates value is already set : %s", candidates)
		}
		err = setVote(scs, []byte(from), []byte(to), (uint64)(math.MaxInt64+i), (uint64)(i))
		if err != nil {
			t.Errorf("failed to setVote : %s", err.Error())
		}
		amount, updateBlockNo, candidates, err = getVote(scs, []byte(from))
		if err != nil {
			t.Errorf("failed to getVote after set : %s", err.Error())
		}
		if amount != (uint64)(math.MaxInt64+i) {
			t.Errorf("invalid amount : %d =/= %d", (uint64)(math.MaxInt64+i), amount)
		}
		if updateBlockNo != (uint64)(i) {
			t.Errorf("invalid block number: %d =/= %d", (uint64)(math.MaxInt64+i), updateBlockNo)
		}
		if !bytes.Equal(candidates, []byte(to)) {
			t.Errorf("invalid candidates : %s =/= %s", string(candidates), to)
		}
	}

}
