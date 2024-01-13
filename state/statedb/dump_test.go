package statedb

import (
	"encoding/json"
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

	dump, err := stateDB.RawDump()
	require.NoError(t, err)

	jsondata, err := json.MarshalIndent(dump, "", "\t")
	require.NoError(t, err)

	require.Equal(t, string(jsondata), `{
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
	err := stateDB.PutState(testAccount, &types.State{
		Balance:  types.NewAmount(1, types.Aergo).Bytes(),
		Nonce:    1,
		CodeHash: common.Hasher([]byte(code)),
	})
	require.NoError(t, err, "failed to put state")

	err = stateDB.Update()
	require.NoError(t, err, "failed to update")
	err = stateDB.Commit()
	require.NoError(t, err, "failed to commit")

	dump, err := stateDB.RawDump()
	require.NoError(t, err)

	jsondata, err := json.MarshalIndent(dump, "", "\t")
	require.NoError(t, err)

	require.Equal(t, string(jsondata), `{
	"accounts": {
		"9RhQjznbYXqMQG1GmuYSsvoCYe5bnCPZCTnT6ZvohkxN": {
			"code": "",
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
