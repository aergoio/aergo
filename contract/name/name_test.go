package name

import (
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB

func initTest(t *testing.T) {
	sdb = state.NewChainStateDB()
	sdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()

	err := sdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
}

func deinitTest() {
	sdb.Close()
	os.RemoveAll("test")
}
func TestName(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name := "AB1234567890"
	owner := types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	buyer := types.ToAddress("AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay")

	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	tx := &types.TxBody{Account: owner, Payload: buildNamePayload(name, 'c', nil)}
	tx.Recipient = []byte(types.AergoName)

	err = CreateName(scs, tx)
	assert.NoError(t, err, "create name")
	err = CreateName(scs, tx)
	assert.Error(t, err, "same name")
	ret := getAddress(scs, []byte(name))
	assert.Equal(t, owner, ret, "registed owner")

	tx.Payload = buildNamePayload(name, 'u', buyer)
	err = UpdateName(scs, tx)
	assert.NoError(t, err, "update name")
	ret = getAddress(scs, []byte(name))
	assert.Equal(t, buyer, ret, "registed owner")
}

func TestNameRecursive(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name1 := "AB1234567890"
	name2 := "1234567890CD"
	owner := types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	buyer := types.ToAddress("AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay")

	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	tx := &types.TxBody{Account: owner, Payload: buildNamePayload(name1, 'c', nil)}
	err = CreateName(scs, tx)
	assert.NoError(t, err, "create name")

	tx.Account = []byte(name1)
	tx.Recipient = []byte(types.AergoName)
	tx.Payload = buildNamePayload(name2, 'c', nil)
	err = CreateName(scs, tx)
	assert.NoError(t, err, "redirect name")

	ret := getAddress(scs, []byte(name2))
	assert.Equal(t, owner, ret, "registed owner")
	name1Owner := GetOwner(scs, []byte(name1))
	t.Logf("name1 owner is %s", types.EncodeAddress(name1Owner.Address))
	assert.Equal(t, owner, name1Owner.Address, "check registed pubkey owner")
	name2Owner := GetOwner(scs, []byte(name2))
	t.Logf("name2 owner is %s", string(name2Owner.Address))
	assert.Equal(t, []byte(name1), name2Owner.Address, "check registed named owner")

	tx.Payload = buildNamePayload(name1, 'u', buyer)
	err = UpdateName(scs, tx)
	assert.NoError(t, err, "update name")
	ret = getAddress(scs, []byte(name1))
	assert.Equal(t, buyer, ret, "registed owner")
}
func TestNameNil(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name1 := "AB1234567890"
	name2 := "1234567890CD"

	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	tx := &types.TxBody{Account: []byte(name1), Payload: buildNamePayload(name2, 'c', nil)}
	err = CreateName(scs, tx)
	assert.NoError(t, err, "create name")
}

func buildNamePayload(name string, operation byte, buyer []byte) []byte {
	payload := []byte{operation}
	payload = append(payload, []byte(name)...)
	if payload[0] == 'u' {
		payload = append(payload, ',')
		payload = append(payload, buyer...)
	}
	return payload
}
