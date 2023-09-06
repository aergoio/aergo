package state

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	testAccount = types.ToAccountID([]byte("test_address"))
	//testRoot, _ = enc.ToBytes("5eGvHsNc5526JBqd8FhKFrtti2fT7xiCyB6rJXt9egFc")
	testRoot   = []byte{0xde, 0xf0, 0x85, 0x93, 0x70, 0x51, 0x4d, 0x51, 0x36, 0x82, 0x9e, 0xeb, 0x4a, 0xd1, 0x6, 0x57, 0x7c, 0xd1, 0xc8, 0x52, 0xc, 0xcb, 0x74, 0xb2, 0xa6, 0x4b, 0xf0, 0x34, 0xc6, 0xf4, 0x5d, 0x80}
	testStates = []types.State{
		{Nonce: 1, Balance: new(big.Int).SetUint64(100).Bytes()},
		{Nonce: 2, Balance: new(big.Int).SetUint64(200).Bytes()},
		{Nonce: 3, Balance: new(big.Int).SetUint64(300).Bytes()},
		{Nonce: 4, Balance: new(big.Int).SetUint64(400).Bytes()},
		{Nonce: 5, Balance: new(big.Int).SetUint64(500).Bytes()},
	}
	//testSecondRoot, _ = enc.ToBytes("GGKZy5XqNPU1VWYspHPwEtm8hnZX2yhcP236ztKf6Pif")
	testSecondRoot   = []byte{0x66, 0xf9, 0x19, 0x2, 0x91, 0xe6, 0xb5, 0x74, 0x3, 0x69, 0x1e, 0x86, 0x87, 0x22, 0x78, 0x1f, 0x4, 0xc3, 0x67, 0x5, 0xf1, 0xb6, 0xce, 0x4b, 0x63, 0x61, 0x6, 0x2c, 0x24, 0xb1, 0xe7, 0xda}
	testSecondStates = []types.State{
		{Nonce: 6, Balance: new(big.Int).SetUint64(600).Bytes()},
		{Nonce: 7, Balance: new(big.Int).SetUint64(700).Bytes()},
		{Nonce: 8, Balance: new(big.Int).SetUint64(800).Bytes()},
	}
)

func stateEquals(expected, actual *types.State) bool {
	return expected.Nonce == actual.Nonce &&
		bytes.Equal(expected.Balance, actual.Balance) &&
		bytes.Equal(expected.CodeHash, actual.CodeHash) &&
		bytes.Equal(expected.StorageRoot, actual.StorageRoot) &&
		expected.SqlRecoveryPoint == actual.SqlRecoveryPoint
}

func TestStateDBGetEmptyState(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// get nil state
	st, err := stateDB.GetState(testAccount)
	if err != nil {
		t.Errorf("failed to get state: %v", err.Error())
	}
	assert.Nil(t, st)

	// get empty state
	st, err = stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.NotNil(t, st)
	assert.Empty(t, st)
}

func TestStateDBPutState(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// put state
	err := stateDB.PutState(testAccount, &testStates[0])
	if err != nil {
		t.Errorf("failed to put state: %v", err.Error())
	}

	// get state
	st, err := stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.NotNil(t, st)
	assert.True(t, stateEquals(&testStates[0], st))
}

func TestStateDBRollback(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// put states
	initialRevision := stateDB.Snapshot()
	for _, v := range testStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	revision := stateDB.Snapshot()
	for _, v := range testSecondStates {
		_ = stateDB.PutState(testAccount, &v)
	}

	// get state
	st, err := stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.NotNil(t, st)
	assert.True(t, stateEquals(&testSecondStates[2], st))

	// rollback to snapshot
	err = stateDB.Rollback(revision)
	if err != nil {
		t.Errorf("failed to rollback: %v", err.Error())
	}
	st, err = stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.NotNil(t, st)
	assert.True(t, stateEquals(&testStates[4], st))

	// rollback to initial revision snapshot
	err = stateDB.Rollback(initialRevision)
	if err != nil {
		t.Errorf("failed to rollback: %v", err.Error())
	}
	st, err = stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.NotNil(t, st)
	assert.Empty(t, st)
}

func TestStateDBUpdateAndCommit(t *testing.T) {
	initTest(t)
	defer deinitTest()

	assert.Nil(t, stateDB.GetRoot())
	for _, v := range testStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	assert.Nil(t, stateDB.GetRoot())

	err := stateDB.Update()
	if err != nil {
		t.Errorf("failed to update: %v", err.Error())
	}
	assert.NotNil(t, stateDB.GetRoot())
	assert.Equal(t, testRoot, stateDB.GetRoot())

	err = stateDB.Commit()
	if err != nil {
		t.Errorf("failed to commit: %v", err.Error())
	}
	assert.Equal(t, testRoot, stateDB.GetRoot())
}

func TestStateDBSetRoot(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// put states
	assert.Nil(t, stateDB.GetRoot())
	for _, v := range testStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	_ = stateDB.Update()
	_ = stateDB.Commit()
	assert.Equal(t, testRoot, stateDB.GetRoot())

	// put additional states
	for _, v := range testSecondStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	_ = stateDB.Update()
	_ = stateDB.Commit()
	assert.Equal(t, testSecondRoot, stateDB.GetRoot())

	// get state
	st, _ := stateDB.GetAccountState(testAccount)
	assert.True(t, stateEquals(&testSecondStates[2], st))

	// set root
	err := stateDB.SetRoot(testRoot)
	if err != nil {
		t.Errorf("failed to set root: %v", err.Error())
	}
	assert.Equal(t, testRoot, stateDB.GetRoot())

	// get state after setting root
	st, err = stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get account state: %v", err.Error())
	}
	assert.True(t, stateEquals(&testStates[4], st))
}

func TestStateDBParallel(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// put states
	assert.Nil(t, stateDB.GetRoot())
	for _, v := range testStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	_ = stateDB.Update()
	_ = stateDB.Commit()
	assert.Equal(t, testRoot, stateDB.GetRoot())

	// put additional states
	for _, v := range testSecondStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	_ = stateDB.Update()
	_ = stateDB.Commit()
	assert.Equal(t, testSecondRoot, stateDB.GetRoot())

	// get state
	st, _ := stateDB.GetAccountState(testAccount)
	assert.True(t, stateEquals(&testSecondStates[2], st))

	// open another statedb with root hash of previous state
	anotherStateDB := chainStateDB.OpenNewStateDB(testRoot)
	assert.Equal(t, testRoot, anotherStateDB.GetRoot())
	assert.Equal(t, testSecondRoot, stateDB.GetRoot())

	// get state from statedb
	st1, err := stateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get state: %v", err.Error())
	}
	assert.True(t, stateEquals(&testSecondStates[2], st1))

	// get state from another statedb
	st2, err := anotherStateDB.GetAccountState(testAccount)
	if err != nil {
		t.Errorf("failed to get state: %v", err.Error())
	}
	assert.True(t, stateEquals(&testStates[4], st2))
}

func TestStateDBMarker(t *testing.T) {
	initTest(t)
	defer deinitTest()
	assert.Nil(t, stateDB.GetRoot())

	for _, v := range testStates {
		_ = stateDB.PutState(testAccount, &v)
	}
	_ = stateDB.Update()
	_ = stateDB.Commit()
	assert.Equal(t, testRoot, stateDB.GetRoot())

	assert.True(t, stateDB.HasMarker(stateDB.GetRoot()))
	assert.False(t, stateDB.HasMarker(testSecondRoot))
	assert.False(t, stateDB.HasMarker([]byte{}))
	assert.False(t, stateDB.HasMarker(nil))
}
