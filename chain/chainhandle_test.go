/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package chain

import (
	"os"
	"testing"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB
var keystore *key.Store

func initTest(t *testing.T) {
	sdb = state.NewChainStateDB()
	keystore = key.NewStore("test")
	sdb.Init("test", nil, false)
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
func makeTestAddress(t *testing.T) []byte {
	addr, err := keystore.CreateKey("test")
	assert.NoError(t, err, "could not create key")
	return addr
}

func signTestAddress(t *testing.T, tx *types.Tx) {
	_, err := keystore.Unlock(tx.GetBody().GetAccount(), "test")
	assert.NoError(t, err, "could not unlock key")
	err = keystore.SignTx(tx)
	assert.NoError(t, err, "could not sign key")
}

func TestErrorInExecuteTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{}

	err := executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx")

	tx.Body = &types.TxBody{}
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx body")

	tx.Body.Account = makeTestAddress(t)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxHasInvalidHash.Error(), "execute tx body with account")

	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxNonceTooLow.Error(), "execute tx body with account")

	tx.Body.Nonce = 1
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with account")
}
