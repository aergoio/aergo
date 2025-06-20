package statedb

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/require"
)

func TestDumpAccount(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// set account state
	for _, v := range testStates {
		err := stateDB.PutState(testAccount, v)
		require.NoError(t, err, "failed to put state")
	}
	err := stateDB.Update()
	require.NoError(t, err, "failed to update")
	err = stateDB.Commit()
	require.NoError(t, err, "failed to commit")

	dump, err := stateDB.Dump()
	require.NoError(t, err)

	require.Equal(t, string(dump), `{
	"accounts": {
		"9RhQjznbYXqMQG1GmuYSsvoCYe5bnCPZCTnT6ZvohkxN": {
			"code": "",
			"state": {
				"balance": "500",
				"codeHash": "",
				"nonce": 5,
				"sqlRecoveryPoint": 0,
				"storageRoot": ""
			},
			"storage": {}
		}
	},
	"root": "G1GECbeFFqSB4WhWCFF8b6tXfc8NCpa2XMqRmgbv7gDD"
}`)
}

func TestDumpContract(t *testing.T) {
	initTest(t)
	defer deinitTest()

	code := "testcode"
	codeHash := common.Hasher([]byte(code))
	// save code
	err := saveData(store, codeHash, []byte(code))
	require.NoError(t, err, "failed to save code")
	// set contract state
	err = stateDB.PutState(testAccount, &types.State{
		Balance:  types.NewAmount(1, types.Aergo).Bytes(),
		Nonce:    1,
		CodeHash: common.Hasher([]byte(code)),
	})
	require.NoError(t, err, "failed to put state")

	err = stateDB.Update()
	require.NoError(t, err, "failed to update")
	err = stateDB.Commit()
	require.NoError(t, err, "failed to commit")

	dump, err := stateDB.Dump()
	require.NoError(t, err)

	require.Equal(t, string(dump), `{
	"accounts": {
		"9RhQjznbYXqMQG1GmuYSsvoCYe5bnCPZCTnT6ZvohkxN": {
			"code": "testcode",
			"state": {
				"balance": "1000000000000000000",
				"codeHash": "6GBoUd26XJnkj6wGs1L6fLg8jhuVXSTVUNWzoXsjeHoh",
				"nonce": 1,
				"sqlRecoveryPoint": 0,
				"storageRoot": ""
			},
			"storage": {}
		}
	},
	"root": "9NfkNYKP6KKZbQ2S3CDkeWqWMfxv2zfNgETCPoqWkDmP"
}`)
}

func TestDumpStorage(t *testing.T) {
	initTest(t)
	defer deinitTest()

	code := "testcode"
	codeHash := common.Hasher([]byte(code))
	// set code
	err := saveData(store, codeHash, []byte(code))
	require.NoError(t, err, "failed to save code")

	// set contract state
	err = stateDB.PutState(testAccount, &types.State{
		Balance:  types.NewAmount(1, types.Aergo).Bytes(),
		Nonce:    1,
		CodeHash: common.Hasher([]byte(code)),
	})
	require.NoError(t, err, "failed to put state")

	// set contract storage
	scs, err := OpenContractStateAccount([]byte("test_address"), stateDB)
	require.NoError(t, err, "failed to open contract state account")
	scs.SetData([]byte("test_storage1"), []byte("test_value1"))
	scs.SetData([]byte("test_storage2"), []byte("test_value2"))
	scs.SetData([]byte("test_storage3"), []byte("test_value3"))
	StageContractState(scs, stateDB)

	err = stateDB.Update()
	require.NoError(t, err, "failed to update")
	err = stateDB.Commit()
	require.NoError(t, err, "failed to commit")

	dump, err := stateDB.Dump()
	require.NoError(t, err)

	require.Equal(t, string(dump), `{
	"accounts": {
		"9RhQjznbYXqMQG1GmuYSsvoCYe5bnCPZCTnT6ZvohkxN": {
			"code": "testcode",
			"state": {
				"balance": "1000000000000000000",
				"codeHash": "6GBoUd26XJnkj6wGs1L6fLg8jhuVXSTVUNWzoXsjeHoh",
				"nonce": 1,
				"sqlRecoveryPoint": 0,
				"storageRoot": "9SVveGGrFXJtoVFFiGpWZ1TmHmKLPnqoYmS7AGxzxgdL"
			},
			"storage": {
				"BVsyGJb6L5qLr8EPcM78Wd5NgZz3cjC1jM2FxpmZFrBm": "Vs5LyU62cV3qLve",
				"ByeMAs5g3t233iEGkEzMYoN4UrnaJPJ6TGNdY2vs9MBg": "Vs5LyU62cV3qLvd",
				"q6MAgzMsY2iJcukzs5x7M9WFmN4bT9AiqzviD1DcSZX": "Vs5LyU62cV3qLvc"
			}
		}
	},
	"root": "9bQx52KdKfMkVdakKEWQLBscgkMiFwd2Zx2hr7u1GYam"
}`)
}

func TestStateDB_RawDumpWith(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// set account state
	for _, v := range testStates {
		err := stateDB.PutState(testAccount, v)
		require.NoError(t, err, "failed to put state")
	}
	stateDB.Update()
	stateDB.Commit()

	W := os.Stdout
	processor := func(idx int64, accountId types.AccountID, account *DumpAccount) error {
		if idx > 0 {
			W.Write([]byte(",\n"))
		}
		accountJson, err := json.MarshalIndent(account, "", "\t")
		if err != nil {
			return err
		}
		W.Write([]byte("\""))
		W.Write([]byte(accountId.String()))
		W.Write([]byte("\":"))
		W.Write(accountJson)
		return nil
	}

	W.Write([]byte("{\n"))
	stateDB.RawDumpWith(DefaultConfig(), processor)
	W.Write([]byte("\n}\n"))
}
