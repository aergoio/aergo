package state

import (
	"bytes"
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var chainStateDB *ChainStateDB
var stateDB *StateDB

func initTest(t *testing.T) {
	chainStateDB = NewChainStateDB()
	_ = chainStateDB.Init(string(db.BadgerImpl), "test", nil, false)
	stateDB = chainStateDB.GetStateDB()
	genesis := types.GetTestGenesis()

	err := chainStateDB.SetGenesis(genesis)
	if err != nil {
		t.Errorf("failed init : %s", err.Error())
	}
}
func deinitTest() {
	_ = chainStateDB.Close()
	_ = os.RemoveAll("test")
}
func TestContractStateCode(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	testBytes := []byte("test_bytes")
	contractState, err := stateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	err = contractState.SetCode(testBytes)
	if err != nil {
		t.Errorf("counld set code to contract state : %s", err.Error())
	}
	res, err := contractState.GetCode()
	if !bytes.Equal(res, testBytes) {
		t.Errorf("different code detected : %s =/= %s", testBytes, string(res))
	}
}

func TestContractStateData(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	testBytes := []byte("test_bytes")
	testKey := []byte("test_key")
	contractState, err := stateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	err = contractState.SetData(testKey, testBytes)
	if err != nil {
		t.Errorf("counld set data to contract state : %s", err.Error())
	}
	res, err := contractState.GetData(testKey)
	if !bytes.Equal(res, testBytes) {
		t.Errorf("different data detected : %s =/= %s", testBytes, string(res))
	}
	err = stateDB.StageContractState(contractState)
	if err != nil {
		t.Errorf("counld commit contract state : %s", err.Error())
	}
}

func TestContractStateEmpty(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	contractState, err := stateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	err = stateDB.StageContractState(contractState)
	if err != nil {
		t.Errorf("counld commit contract state : %s", err.Error())
	}
}

func TestContractStateReOpenData(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	testBytes := []byte("test_bytes")
	testKey := []byte("test_key")
	contractState, err := stateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	err = contractState.SetData(testKey, testBytes)
	if err != nil {
		t.Errorf("counld set data to contract state : %s", err.Error())
	}
	res, err := contractState.GetData(testKey)
	if err != nil {
		t.Errorf("counld set data to contract state : %s", err.Error())
	}
	if !bytes.Equal(res, testBytes) {
		t.Errorf("different data detected : %s =/= %s", testBytes, string(res))
	}
	err = stateDB.StageContractState(contractState)
	if err != nil {
		t.Errorf("counld commit contract state : %s", err.Error())
	}
	//contractState2, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	contractState2, err := stateDB.OpenContractState(types.ToAccountID(testAddress), contractState.State)
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	res2, err := contractState2.GetData(testKey)
	if err != nil {
		t.Errorf("counld not get contract state : %s", err.Error())
	}
	if !bytes.Equal(res2, testBytes) {
		t.Errorf("different data detected : %s =/= %s", testBytes, string(res2))
	}
}

func TestContractStateRollback(t *testing.T) {
	initTest(t)
	defer deinitTest()

	testAddress := []byte("test_address")
	testKey := []byte("test_key")
	contractState, err := stateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}

	// test data
	_ = contractState.SetData(testKey, []byte("1")) // rev 1
	_ = contractState.SetData(testKey, []byte("2")) // rev 2
	res, _ := contractState.GetData(testKey)
	assert.Equal(t, []byte("2"), res)

	// snapshot: rev 2
	revision := contractState.Snapshot()
	t.Log("revision", revision)
	assert.Equal(t, 2, int(revision))

	// test data
	_ = contractState.SetData(testKey, []byte("3")) // rev 3
	_ = contractState.SetData(testKey, []byte("4")) // rev 4
	_ = contractState.SetData(testKey, []byte("5")) // rev 5
	res, _ = contractState.GetData(testKey)
	assert.Equal(t, []byte("5"), res)

	// rollback: rev 2
	contractState.Rollback(revision)
	res, _ = contractState.GetData(testKey)
	assert.Equal(t, []byte("2"), res)

	// rollback to empty: rev 0
	contractState.Rollback(Snapshot(0))
	res, _ = contractState.GetData(testKey)
	assert.Nil(t, res)
}
