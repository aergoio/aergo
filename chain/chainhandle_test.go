/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package chain

import (
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB
var keystore *key.Store

func initTest(t *testing.T, testmode bool) {
	sdb = state.NewChainStateDB()
	tmpdir, _ := ioutil.TempDir("", "test")
	keystore = key.NewStore(tmpdir)
	sdb.Init(string(db.BadgerImpl), tmpdir, nil, testmode)
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
	initTest(t, true)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{}

	err := executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx")

	tx.Body = &types.TxBody{}
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx body")

	tx.Body.Account = makeTestAddress(t)
	tx.Body.Recipient = makeTestAddress(t)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxHasInvalidHash.Error(), "execute tx body with account")

	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrTxNonceTooLow.Error(), "execute tx body with account")

	tx.Body.Nonce = 1
	tx.Body.Amount = math.MaxUint64
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with nonce")

	tx.Body.Amount = types.MaxAER
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with nonce")
}

func TestBasicExecuteTx(t *testing.T) {
	initTest(t, true)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{Body: &types.TxBody{}}

	tx.Body.Account = makeTestAddress(t)
	tx.Body.Recipient = makeTestAddress(t)
	tx.Body.Nonce = 1
	signTestAddress(t, tx)
	err := executeTx(bs, tx, 0, 0)
	assert.NoError(t, err, "execute amount 0")

	tx.Body.Nonce = 2
	tx.Body.Amount = 1000
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.NoError(t, err, "execute amount 1000")

	tx.Body.Nonce = 3
	tx.Body.Amount = 10000
	tx.Body.Recipient = []byte(types.AergoSystem)
	tx.Body.Type = types.TxType_GOVERNANCE
	tx.Body.Payload = []byte{'s'}
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0)
	assert.NoError(t, err, "execute governance type")

}
