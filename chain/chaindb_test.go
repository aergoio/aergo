package chain

import (
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestChainDB_hardforkHeights(t *testing.T) {
	type fields struct {
		cc        consensus.ChainConsensus
		latest    atomic.Value
		bestBlock atomic.Value
		store     db.DB
	}
	tests := []struct {
		name   string
		hardfork string
		wantSuccess bool
	}{
		{ "normal", `{"V2":2000,"V3":30000,"V4":400000}`, true},
		{ "wrong", `{"A":2000,"B3A":30000,"V4":400000}`, false},
		{ "wrong2", `{"A2":2000,"B3":30000,"V4":400000}`, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdb := &ChainDB{
				store:     DBStub{tt.hardfork},
			}
			actual := cdb.hardforkHeights()
			if tt.wantSuccess {
				assert.Equal(t, 3,len(actual),  "hardforkHeights()")
				assert.Equal(t, uint64(2000), actual["2"], "hardforkHeights() : 2")
				assert.Equal(t, uint64(30000), actual["3"], "hardforkHeights() : 3")
				assert.Equal(t, uint64(400000), actual["4"], "hardforkHeights() : 4")
				assert.Equal(t, uint64(0), actual["5"], "hardforkHeights() : 5")
			} else {
				assert.Equal(t, 0, len(actual), "hardforkHeights()")
			}
		})
	}
}

type DBStub struct {
	hardfork string
}

func (m DBStub) Type() string {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Set(key, value []byte) {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Delete(key []byte) {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Get(key []byte) []byte {
	if string(key) == "hardfork" {
		return []byte(m.hardfork)
	}
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Exist(key []byte) bool {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Iterator(start, end []byte) db.Iterator {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) NewTx() db.Transaction {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) NewBulk() db.Bulk {
	//TODO implement me
	panic("implement me")
}

func (m DBStub) Close() {
	//TODO implement me
	panic("implement me")
}
