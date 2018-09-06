/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/aergoio/aergo/state"

	"github.com/aergoio/aergo/types"
)

var sdb *state.ChainStateDB

func initTest(t *testing.T) {
	sdb = state.NewStateDB()
	sdb.Init("test")
	genesisBlock := &types.Block{}
	err := sdb.SetGenesis(genesisBlock)
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
		err := updateVoteResult(scs, []byte(to), (int64)(i*i), (uint64)(i))
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
		to := fmt.Sprintf("to%d", i)
		voting, updateBlockNo, err := getVote(scs, []byte(from), []byte(to))
		if err != nil {
			t.Errorf("failed to getVote : %s", err.Error())
		}
		if voting != 0 {
			t.Errorf("new voting value is already set : %d", voting)
		}
		if updateBlockNo != 0 {
			t.Errorf("new block number value is already set : %d", updateBlockNo)
		}
		err = setVote(scs, []byte(from), []byte(to), (uint64)(math.MaxInt64+i), (uint64)(i))
		if err != nil {
			t.Errorf("failed to setVote : %s", err.Error())
		}
		voting, updateBlockNo, err = getVote(scs, []byte(from), []byte(to))
		if err != nil {
			t.Errorf("failed to getVote after set : %s", err.Error())
		}
		if voting != (uint64)(math.MaxInt64+i) {
			t.Errorf("invalid voting value : %d =/= %d", (uint64)(math.MaxInt64+i), voting)
		}
		if updateBlockNo != (uint64)(i) {
			t.Errorf("invalid block number: %d =/= %d", i, updateBlockNo)
		}
	}

}
