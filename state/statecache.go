package state

import (
	"sort"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
)

var (
	emptyBufferEntry = bufferEntry{}
)

type bufferEntry struct {
	key       types.HashID
	dataHash  types.HashID
	dataBytes []byte
}

func newBufferEntry(key types.HashID, data []byte) bufferEntry {
	return bufferEntry{
		key:       key,
		dataHash:  types.GetHashID(data),
		dataBytes: data,
	}
}

type stateBuffer struct {
	lock    sync.RWMutex
	entries []bufferEntry
	indexes map[types.HashID]int
}

func newStateBuffer() *stateBuffer {
	return &stateBuffer{
		entries: []bufferEntry{},
		indexes: map[types.HashID]int{},
	}
}

func (buffer *stateBuffer) get(key types.HashID) *bufferEntry {
	buffer.lock.RLock()
	defer buffer.lock.RUnlock()
	if index, ok := buffer.indexes[key]; ok {
		return &buffer.entries[index]
	}
	return nil
}

func (buffer *stateBuffer) puts(ets ...bufferEntry) {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	for _, v := range ets {
		buffer.entries = append(buffer.entries, v)
		buffer.indexes[v.key] = buffer.snapshot()
	}
}

func (buffer *stateBuffer) snapshot() int {
	// TODO: last index of entries
	return len(buffer.entries) - 1
}

func (buffer *stateBuffer) revert(snapshot int) {
	// TODO: revert entries and indexes
}

func (buffer *stateBuffer) isEmpty() bool {
	return len(buffer.entries) == 0
}

func (buffer *stateBuffer) export() ([][]byte, [][]byte) {
	buffer.lock.RLock()
	defer buffer.lock.RUnlock()
	size := len(buffer.indexes)
	bufs := make([]bufferEntry, 0, size)
	for _, v := range buffer.indexes {
		bufs = append(bufs, buffer.entries[v])
	}
	sort.Slice(bufs, func(i, j int) bool {
		return -1 == (bufs[i].key).Compare(bufs[j].key)
	})
	keys := make([][]byte, size)
	vals := make([][]byte, size)
	for i, et := range bufs {
		keys[i] = append(keys[i], et.key[:]...)
		vals[i] = append(vals[i], et.dataHash[:]...)
	}
	return keys, vals
}

func (buffer *stateBuffer) commit(store *db.DB) {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	dbtx := (*store).NewTx()
	for _, v := range buffer.indexes {
		et := buffer.entries[v]
		dbtx.Set(et.dataHash[:], et.dataBytes)
	}
	dbtx.Commit()
}
