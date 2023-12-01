package ethdb

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testPath = "./teststate"
)

func makedbPath(dbType string) string {
	return testPath + "/" + dbType
}

func TestRawDB(t *testing.T) {
	os.MkdirAll(testPath, 0775)
	defer os.RemoveAll(testPath)

	for _, test := range []struct {
		dbName string
	}{
		{"memorydb"}, {"leveldb"}, {"pebbledb"},
	} {
		db, err := NewDB(makedbPath(test.dbName), test.dbName)
		require.NoError(t, err)
		res, _ := db.Store.Get([]byte("foo"))
		if res != nil {
			t.Errorf("fetching non-existant yields non nil item")
			t.Log(string(res))
		}
		db.Store.Put([]byte("foo"), []byte("bar"))
		res, _ = db.Store.Get([]byte("foo"))
		if !bytes.Equal(res, []byte("bar")) {
			t.Errorf("retrieved value does not match")
		}
		db.Close()
	}
}

func TestTrie(t *testing.T) {
	dbName := "memorydb"
	db, err := NewDB(makedbPath(dbName), dbName)
	require.NoError(t, err)
	defer db.Close()

}
