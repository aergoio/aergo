package statedb

import (
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/stretchr/testify/assert"
)

var (
	testKey  = []byte("test_key")
	testData = []byte("test_data")
	testOver = []byte("test_over")
)

var (
	store   db.DB
	stateDB *StateDB
)

func initTest(t *testing.T) {
	dbPath := common.PathMkdirAll("test", StateName)
	store = db.NewDB(db.ImplType(db.BadgerImpl), dbPath)
	stateDB = NewStateDB(store, nil, false)
}
func deinitTest() {
	store.Close()
	stateDB = nil
	_ = os.RemoveAll("test")
}
func TestStateDataBasic(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// save data
	if err := saveData(store, testKey, testData); err != nil {
		t.Errorf("failed to save data: %v", err.Error())
	}

	// load data
	data := []byte{}
	if err := loadData(store, testKey, &data); err != nil {
		t.Errorf("failed to load data: %v", err.Error())
	}
	assert.NotNil(t, data)
	assert.Equal(t, testData, data)
}

func TestStateDataNil(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// load data before saving
	var data interface{}
	assert.Nil(t, data)
	if err := loadData(store, testKey, &data); err != nil {
		t.Errorf("failed to load data: %v", err.Error())
	}
	assert.Nil(t, data)
}

func TestStateDataEmpty(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// save empty data
	var testEmpty []byte
	if err := saveData(store, testKey, testEmpty); err != nil {
		t.Errorf("failed to save nil data: %v", err.Error())
	}

	// load empty data
	data := []byte{}
	if err := loadData(store, testKey, &data); err != nil {
		t.Errorf("failed to load data: %v", err.Error())
	}
	assert.NotNil(t, data)
	assert.Empty(t, data)
}

func TestStateDataOverwrite(t *testing.T) {
	initTest(t)
	defer deinitTest()

	// save data
	if err := saveData(store, testKey, testData); err != nil {
		t.Errorf("failed to save data: %v", err.Error())
	}

	// save another data to same key
	if err := saveData(store, testKey, testOver); err != nil {
		t.Errorf("failed to overwrite data: %v", err.Error())
	}

	// load data
	data := []byte{}
	if err := loadData(store, testKey, &data); err != nil {
		t.Errorf("failed to load data: %v", err.Error())
	}
	assert.NotNil(t, data)
	assert.Equal(t, testOver, data)
}
