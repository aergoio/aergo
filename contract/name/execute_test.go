package name

import (
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestExcuteNameTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txBody := &types.TxBody{}
	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	txBody.Account = types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	txBody.Recipient = []byte(types.AergoName)

	name := "AB1234567890"
	txBody.Payload = buildNamePayload(name, 'c', nil)
	err = ExecuteNameTx(scs, txBody)
	assert.NoError(t, err, "execute name tx")
	ret := GetAddress(scs, []byte(name))
	assert.Equal(t, txBody.Account, ret, "pubkey address")

}
