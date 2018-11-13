package state

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

var (
	testAccount = types.ToAccountID([]byte("test_address"))
	testRoot, _ = enc.ToBytes("Fuep8wrfFDjV1tQRJThT5Ukdb3zgFiQdB2sVZnox232b")
	testStates  = []types.State{
		types.State{Nonce: 1, Balance: 100},
		types.State{Nonce: 2, Balance: 200},
		types.State{Nonce: 3, Balance: 300},
		types.State{Nonce: 4, Balance: 400},
		types.State{Nonce: 5, Balance: 500},
	}
	testSecondRoot, _ = enc.ToBytes("24yWBBoCZB9dmGX4fLSbQuHpQdNWPxoQRxgG14KEz8Du")
	testSecondStates  = []types.State{
		types.State{Nonce: 6, Balance: 600},
		types.State{Nonce: 7, Balance: 700},
		types.State{Nonce: 8, Balance: 800},
	}
)

func stateEquals(expected, actual *types.State) bool {
	return expected.Nonce == actual.Nonce &&
		expected.Balance == actual.Balance &&
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
