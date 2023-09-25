package name

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var sdb *state.ChainStateDB
var block *types.Block

func initTest(t *testing.T) {
	genesis := types.GetTestGenesis()
	sdb = state.NewChainStateDB()
	sdb.Init(string(db.BadgerImpl), "test", genesis.Block(), false)
	err := sdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	block = genesis.Block()
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
	buyer := "AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay"

	tx := &types.TxBody{Account: owner, Payload: buildNamePayload(name, types.NameCreate, "")}
	tx.Recipient = []byte(types.AergoName)

	sender, _ := sdb.GetStateDB().GetAccountStateV(tx.Account)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(tx.Recipient)
	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)
	systemcs := openSystemContractState(t, bs)

	err := CreateName(scs, tx, sender, receiver, name)
	assert.NoError(t, err, "create name")

	scs = nextBlockContractState(t, bs, scs)
	_, err = ValidateNameTx(tx, sender, scs, systemcs)
	assert.Error(t, err, "same name")

	ret := getAddress(scs, []byte(name))
	assert.Equal(t, owner, ret, "registed owner")

	tx.Payload = buildNamePayload(name, types.NameUpdate, buyer)
	err = UpdateName(bs, scs, tx, sender, receiver, name, buyer)
	assert.NoError(t, err, "update name")

	scs = nextBlockContractState(t, bs, scs)

	ret = getAddress(scs, []byte(name))
	assert.Equal(t, buyer, types.EncodeAddress(ret), "registed owner")
}

func TestNameRecursive(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name1 := "AB1234567890"
	name2 := "1234567890CD"
	owner := types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	buyer := "AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay"

	tx := &types.TxBody{Account: owner, Recipient: []byte(types.AergoName), Payload: buildNamePayload(name1, types.NameCreate, "")}

	sender, _ := sdb.GetStateDB().GetAccountStateV(tx.Account)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(tx.Recipient)
	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)
	err := CreateName(scs, tx, sender, receiver, name1)
	assert.NoError(t, err, "create name")

	tx.Account = []byte(name1)
	tx.Recipient = []byte(types.AergoName)
	tx.Payload = buildNamePayload(name2, types.NameCreate, "")

	scs = nextBlockContractState(t, bs, scs)
	err = CreateName(scs, tx, sender, receiver, name2)
	assert.NoError(t, err, "redirect name")

	scs = nextBlockContractState(t, bs, scs)
	ret := getAddress(scs, []byte(name2))
	assert.Equal(t, owner, ret, "registed owner")
	name1Owner := GetOwner(scs, []byte(name1))
	t.Logf("name1 owner is %s", types.EncodeAddress(name1Owner))
	assert.Equal(t, owner, name1Owner, "check registed pubkey owner")
	name2Owner := GetOwner(scs, []byte(name2))
	t.Logf("name2 owner is %s", types.EncodeAddress(name2Owner))
	assert.Equal(t, owner, name2Owner, "check registed named owner")

	tx.Payload = buildNamePayload(name1, types.NameUpdate, buyer)

	err = UpdateName(bs, scs, tx, sender, receiver, name1, buyer)
	assert.NoError(t, err, "update name")
	scs = nextBlockContractState(t, bs, scs)
	ret = getAddress(scs, []byte(name1))
	assert.Equal(t, buyer, types.EncodeAddress(ret), "registed owner")
}

func TestNameNil(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name1 := "AB1234567890"
	name2 := "1234567890CD"

	scs, err := sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	tx := &types.TxBody{Account: []byte(name1), Payload: buildNamePayload(name2, types.NameCreate, "")}
	sender, _ := sdb.GetStateDB().GetAccountStateV(tx.Account)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(tx.Recipient)

	err = CreateName(scs, tx, sender, receiver, name2)
	assert.NoError(t, err, "create name")
}

func TestNameSetContractOwner(t *testing.T) {
	initTest(t)
	defer deinitTest()
	//name := "AB1234567890"
	requester := types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	//ownerAddr := types.ToAddress("AmNExAm3zWL8j4xXDkx2UBCMoLoTgX13K3M4GM8mLPhRCq4KuLeR")
	tx := &types.TxBody{
		Account: requester,
		Payload: []byte(`{"Name":"v1setOwner","Args":["AmNExAm3zWL8j4xXDkx2UBCMoLoTgX13K3M4GM8mLPhRCq4KuLeR"]}`),
	}
	tx.Recipient = []byte(types.AergoName)

	sender, _ := sdb.GetStateDB().GetAccountStateV(tx.Account)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(tx.Recipient)
	//owner, _ := sdb.GetStateDB().GetAccountStateV(ownerAddr)

	receiver.AddBalance(big.NewInt(1000))

	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)
	//systemcs := openSystemContractState(t, bs)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0), ForkVersion: 0}
	_, err := ExecuteNameTx(bs, scs, tx, sender, receiver, blockInfo)
	assert.NoError(t, err, "execute name")
	assert.Equal(t, big.NewInt(0), receiver.Balance(), "check remain")
}

func buildNamePayload(name string, operation string, buyer string) []byte {
	var ci types.CallInfo
	ci.Name = operation
	if buyer != "" {
		ci.Args = append(ci.Args, name, buyer)
	} else {
		ci.Args = append(ci.Args, name)
	}
	payload, err := json.Marshal(ci)
	if err != nil {
		return nil
	}
	return payload
}

func TestNameMap(t *testing.T) {
	initTest(t)
	defer deinitTest()
	name1 := "AB1234567890"
	name2 := "1234567890CD"
	owner := types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	destination := types.ToAddress("AmhhvY54uW2ZbcyeADPs6MKtj6CrTYcuhHpxn1VxPST1usEY1rDm")

	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)

	testNameMap := &NameMap{Owner: owner, Destination: destination, Version: 1}
	err := setNameMap(scs, []byte(name1), testNameMap)
	assert.NoError(t, err, "error return in setNameMap")

	err = setNameMap(scs, []byte(name2), testNameMap)
	assert.NoError(t, err, "error return in setNameMap")

	res := getNameMap(scs, []byte(name1), false)
	assert.Equal(t, testNameMap.Owner, res.Owner, "Owner")
	assert.Equal(t, testNameMap.Destination, res.Destination, "Destination")
	assert.Equal(t, testNameMap.Version, res.Version, "Version")

	res = getNameMap(scs, []byte(name2), false)
	assert.Equal(t, testNameMap.Owner, res.Owner, "Owner")
	assert.Equal(t, testNameMap.Destination, res.Destination, "Destination")
	assert.Equal(t, testNameMap.Version, res.Version, "Version")

	resOwner := getOwner(scs, []byte(name1), false)
	assert.Equal(t, testNameMap.Owner, resOwner, "getOwner")

	scs = nextBlockContractState(t, bs, scs)

	res = getNameMap(scs, []byte(name1), true)
	assert.Equal(t, testNameMap.Owner, res.Owner, "Owner")
	assert.Equal(t, testNameMap.Destination, res.Destination, "Destination")
	assert.Equal(t, testNameMap.Version, res.Version, "Version")

	res = getNameMap(scs, []byte(name2), true)
	assert.Equal(t, testNameMap.Owner, res.Owner, "Owner")
	assert.Equal(t, testNameMap.Destination, res.Destination, "Destination")
	assert.Equal(t, testNameMap.Version, res.Version, "Version")

	resOwner = GetOwner(scs, []byte(name1))
	assert.Equal(t, testNameMap.Owner, resOwner, "GetOwner")
	resAddr := getAddress(scs, []byte(name1))
	assert.Equal(t, testNameMap.Destination, resAddr, "getAddress")

}
