package state

import (
	"bytes"
	"os"
	"testing"

	"github.com/aergoio/aergo/types"
)

var chainStateDB *ChainStateDB

func initTest(t *testing.T) {
	chainStateDB = NewStateDB()
	chainStateDB.Init("test")
	genesisBlock := &types.Genesis{}
	genesisBlock.Block = &types.Block{}
	err := chainStateDB.SetGenesis(genesisBlock)
	if err != nil {
		t.Errorf("failed init : %s", err.Error())
	}
}
func deinitTest() {
	chainStateDB.Close()
	os.RemoveAll("test")
}
func TestContractStateCode(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	testBytes := []byte("test_bytes")
	contractState, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
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
	contractState, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
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
	err = chainStateDB.CommitContractState(contractState)
	if err != nil {
		t.Errorf("counld commit contract state : %s", err.Error())
	}
}

func TestContractStateEmpty(t *testing.T) {
	initTest(t)
	defer deinitTest()
	testAddress := []byte("test_address")
	contractState, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	if err != nil {
		t.Errorf("counld not open contract state : %s", err.Error())
	}
	err = chainStateDB.CommitContractState(contractState)
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
	contractState, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
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
	err = chainStateDB.CommitContractState(contractState)
	if err != nil {
		t.Errorf("counld commit contract state : %s", err.Error())
	}
	//contractState2, err := chainStateDB.OpenContractStateAccount(types.ToAccountID(testAddress))
	contractState2, err := chainStateDB.OpenContractState(contractState.State)
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
