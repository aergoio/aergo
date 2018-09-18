package state

import (
	"sort"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
)

var (
	emptyCacheEntry = cacheEntry{}
)

type cacheEntry struct {
	key       types.HashID
	dataHash  types.HashID
	dataBytes []byte
}

func newCacheEntry(key types.HashID, data []byte) cacheEntry {
	return cacheEntry{
		key:       key,
		dataHash:  types.GetHashID(data),
		dataBytes: data,
	}
}

type stateCaches struct {
	lock    sync.RWMutex
	entries []cacheEntry
	indexes map[types.HashID]int
}

func newStateCaches() *stateCaches {
	return &stateCaches{
		entries: []cacheEntry{},
		indexes: map[types.HashID]int{},
	}
}

func (caches *stateCaches) get(key types.HashID) *cacheEntry {
	caches.lock.RLock()
	defer caches.lock.RUnlock()
	if index, ok := caches.indexes[key]; ok {
		return &caches.entries[index]
	}
	return nil
}

func (caches *stateCaches) puts(ets ...cacheEntry) {
	caches.lock.Lock()
	defer caches.lock.Unlock()
	for _, v := range ets {
		caches.entries = append(caches.entries, v)
		caches.indexes[v.key] = caches.snapshot()
	}
}

func (caches *stateCaches) snapshot() int {
	// TODO: last index of entries
	return len(caches.entries) - 1
}

func (caches *stateCaches) revert(snapshot int) {
	// TODO: revert entries and indexes
}

func (caches *stateCaches) isEmpty() bool {
	return len(caches.entries) == 0
}

func (caches *stateCaches) export() ([][]byte, [][]byte) {
	caches.lock.RLock()
	defer caches.lock.RUnlock()
	size := len(caches.indexes)
	bufs := make([]cacheEntry, 0, size)
	for _, v := range caches.indexes {
		bufs = append(bufs, caches.entries[v])
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

func (caches *stateCaches) commit(store *db.DB) {
	caches.lock.Lock()
	defer caches.lock.Unlock()
	dbtx := (*store).NewTx(true)
	for _, v := range caches.indexes {
		et := caches.entries[v]
		dbtx.Set(et.dataHash[:], et.dataBytes)
	}
	dbtx.Commit()
}
