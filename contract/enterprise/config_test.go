package enterprise

import (
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var cdb *state.ChainStateDB
var sdb *state.StateDB

func initTest(t *testing.T) (*state.ContractState, *state.V, *state.V) {
	cdb = state.NewChainStateDB()
	cdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()
	sdb = cdb.OpenNewStateDB(cdb.GetRoot())
	err := cdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"

	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")
	sender, err := sdb.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV([]byte(types.AergoEnterprise))
	assert.NoError(t, err, "could not get test address state")
	return scs, sender, receiver
}

func deinitTest() {
	cdb.Close()
	os.RemoveAll("test")
}

func TestSetGetConf(t *testing.T) {
	scs, _, _ := initTest(t)
	defer deinitTest()
	testConf := &Conf{On: true, Values: []string{"abc", "def", "ghi"}}
	retConf, err := getConf(scs, []byte("test"))
	assert.NoError(t, err, "could not get test conf")
	assert.Nil(t, retConf, "not set yet")
	err = setConf(scs, []byte("test"), testConf)
	assert.NoError(t, err, "could not set test conf")
	retConf, err = getConf(scs, []byte("test"))
	assert.NoError(t, err, "could not get test conf")
	assert.Equal(t, testConf.Values, retConf.Values, "check values")
	assert.Equal(t, testConf.On, retConf.On, "check on")

	testConf2 := &Conf{On: false, Values: []string{"1", "22", "333"}}
	err = setConf(scs, []byte("test"), testConf2)
	assert.NoError(t, err, "could not set test conf")
	retConf2, err := getConf(scs, []byte("test"))
	assert.NoError(t, err, "could not get test conf")
	assert.Equal(t, testConf2.Values, retConf2.Values, "check values")
	assert.Equal(t, testConf2.On, retConf2.On, "check on")
}
