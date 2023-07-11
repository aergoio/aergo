/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package chain

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/account/key"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB
var keystore *key.Store
var chainID []byte

func initTest(t *testing.T, testmode bool) {
	sdb = state.NewChainStateDB()
	tmpdir, _ := ioutil.TempDir("", "test")
	keystore = key.NewStore(tmpdir, 0)
	sdb.Init(string(db.BadgerImpl), tmpdir, nil, testmode)
	genesis := types.GetTestGenesis()
	chainID = genesis.Block().GetHeader().ChainID

	err := sdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	types.InitGovernance("dpos", true)
	system.InitGovernance("dpos")

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

func newTestBlockInfo(chainID []byte) *types.BlockHeaderInfo {
	return types.NewBlockHeaderInfo(
		&types.Block{
			Header: &types.BlockHeader{
				ChainID:       chainID,
				BlockNo:       0,
				Timestamp:     0,
				PrevBlockHash: nil,
			},
		},
	)
}

func TestErrorInExecuteTx(t *testing.T) {
	initTest(t, true)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{}

	err := executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrTxFormatInvalid.Error(), "execute empty tx")

	tx.Body = &types.TxBody{}

	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrTxInvalidChainIdHash.Error(), "execute empty tx body")

	tx.Body.ChainIdHash = common.Hasher(chainID)
	tx.Body.Account = makeTestAddress(t)
	tx.Body.Recipient = makeTestAddress(t)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrTxHasInvalidHash.Error(), "execute tx body with account")

	signTestAddress(t, tx)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrTxNonceTooLow.Error(), "execute tx body with account")

	tx.Body.Nonce = 1
	tx.Body.Amount = new(big.Int).Add(types.StakingMinimum, types.StakingMinimum).Bytes()
	signTestAddress(t, tx)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with nonce")

	tx.Body.Amount = types.MaxAER.Bytes()
	signTestAddress(t, tx)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.EqualError(t, err, types.ErrInsufficientBalance.Error(), "execute tx body with nonce")
}

func TestBasicExecuteTx(t *testing.T) {
	initTest(t, true)
	defer deinitTest()
	bs := state.NewBlockState(sdb.GetStateDB())

	tx := &types.Tx{Body: &types.TxBody{}}

	tx.Body.ChainIdHash = common.Hasher(chainID)
	tx.Body.Account = makeTestAddress(t)
	tx.Body.Recipient = makeTestAddress(t)
	tx.Body.Nonce = 1
	signTestAddress(t, tx)
	err := executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.NoError(t, err, "execute amount 0")

	tx.Body.Nonce = 2
	tx.Body.Amount = new(big.Int).SetUint64(1000).Bytes()
	signTestAddress(t, tx)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.NoError(t, err, "execute amount 1000")

	tx.Body.Nonce = 3
	tx.Body.Amount = (new(big.Int).Add(types.StakingMinimum, new(big.Int).SetUint64(1))).Bytes()
	tx.Body.Amount = types.StakingMinimum.Bytes()
	tx.Body.Recipient = []byte(types.AergoSystem)
	tx.Body.Type = types.TxType_GOVERNANCE
	tx.Body.Payload = []byte(`{"Name":"v1stake"}`)
	signTestAddress(t, tx)
	err = executeTx(nil, nil, bs, types.NewTransaction(tx), newTestBlockInfo(chainID), contract.ChainService)
	assert.NoError(t, err, "execute governance type")
}
