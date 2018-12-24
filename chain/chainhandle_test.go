/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package chain

import (
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/contract"
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

	err := sdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}

	coinbaseFee, _ = genesis.ID.GetCoinbaseFee()
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
	err = keystore.SignTx(tx, nil)
	assert.NoError(t, err, "could not sign key")
}

func TestErrorInExecuteTx(t *testing.T) {
	initTest(t, true)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{}

	err := executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx")

	tx.Body = &types.TxBody{}
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx body")

	tx.Body.Account = makeTestAddress(t)
	tx.Body.Recipient = makeTestAddress(t)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.EqualError(t, err, types.ErrTxHasInvalidHash.Error(), "execute tx body with account")

	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.EqualError(t, err, types.ErrTxNonceTooLow.Error(), "execute tx body with account")

	tx.Body.Nonce = 1
	tx.Body.Amount = new(big.Int).SetUint64(math.MaxUint64).Bytes()
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with nonce")

	tx.Body.Amount = types.MaxAER.Bytes()
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
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
	err := executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.NoError(t, err, "execute amount 0")

	tx.Body.Nonce = 2
	tx.Body.Amount = new(big.Int).SetUint64(1000).Bytes()
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.NoError(t, err, "execute amount 1000")

	tx.Body.Nonce = 3
	tx.Body.Amount = (new(big.Int).Add(types.StakingMinimum, new(big.Int).SetUint64(1))).Bytes()
	tx.Body.Recipient = []byte(types.AergoSystem)
	tx.Body.Type = types.TxType_GOVERNANCE
	tx.Body.Payload = []byte{'s'}
	signTestAddress(t, tx)
	err = executeTx(bs, tx, 0, 0, nil, contract.ChainService)
	assert.NoError(t, err, "execute governance type")

}
