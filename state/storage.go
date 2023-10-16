package state

import (
	"bytes"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/pkg/trie"
	"github.com/aergoio/aergo/v2/types"
)

var (
	checkpointKey = types.ToHashID([]byte("checkpoint"))
)

type storageCache struct {
	lock     sync.RWMutex
	storages map[types.AccountID]*bufferedStorage
}

func newStorageCache() *storageCache {
	return &storageCache{
		storages: map[types.AccountID]*bufferedStorage{},
	}
}

func (cache *storageCache) snapshot() map[types.AccountID]int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	result := make(map[types.AccountID]int)
	for aid, bs := range cache.storages {
		result[aid] = bs.buffer.snapshot()
	}
	return result
}

func (cache *storageCache) rollback(snap map[types.AccountID]int) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	for aid, bs := range cache.storages {
		if rev, ok := snap[aid]; ok {
			if err := bs.buffer.rollback(rev); err != nil {
				return err
			}
		} else {
			delete(cache.storages, aid)
		}
	}
	return nil
}

func (cache *storageCache) get(key types.AccountID) *bufferedStorage {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	if storage, ok := cache.storages[key]; ok && storage != nil {
		return storage
	}
	return nil
}
func (cache *storageCache) put(key types.AccountID, storage *bufferedStorage) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.storages[key] = storage
}

type bufferedStorage struct {
	buffer *stateBuffer
	trie   *trie.Trie
	dirty  bool
}

func newBufferedStorage(root []byte, store db.DB) *bufferedStorage {
	return &bufferedStorage{
		buffer: newStateBuffer(),
		trie:   trie.NewTrie(root, common.Hasher, store),
		dirty:  false,
	}
}

func (storage *bufferedStorage) has(key types.HashID, lookupTrie bool) bool {
	if storage.buffer.has(key) {
		return true
	}
	if lookupTrie {
		if buf, _ := storage.trie.Get(key.Bytes()); buf != nil {
			return true
		}
	}
	return false
}
func (storage *bufferedStorage) get(key types.HashID) entry {
	return storage.buffer.get(key)
}
func (storage *bufferedStorage) put(et entry) {
	storage.buffer.put(et)
}

func (storage *bufferedStorage) checkpoint(revision int) {
	storage.buffer.put(newMetaEntry(checkpointKey, revision))
}

func (storage *bufferedStorage) rollback(revision int) {
	checkpoints, ok := storage.buffer.indexes[checkpointKey]
	if !ok {
		// do nothing
		return
	}
	it := checkpoints.iter()
	for rev := it(); rev >= 0; rev = it() {
		et := storage.buffer.entries[rev]
		if et == nil {
			continue
		}
		me, ok := et.(*metaEntry)
		if !ok {
			continue
		}
		val, ok := me.Value().(int)
		if !ok {
			continue
		}
		if val < revision {
			break
		}
		storage.buffer.rollback(rev)
	}
}

func (storage *bufferedStorage) update() error {
	before := storage.trie.Root
	if err := storage.buffer.updateTrie(storage.trie); err != nil {
		return err
	}
	if !bytes.Equal(before, storage.trie.Root) {
		logger.Debug().Str("before", enc.ToString(before)).
			Str("after", enc.ToString(storage.trie.Root)).Msg("Changed storage trie root")
		storage.dirty = true
	}
	return nil
}

func (storage *bufferedStorage) isDirty() bool {
	return storage.dirty
}

func (storage *bufferedStorage) stage(txn trie.DbTx) error {
	storage.trie.StageUpdates(txn)
	if err := storage.buffer.stage(txn); err != nil {
		return err
	}
	if err := storage.buffer.reset(); err != nil {
		return err
	}
	return nil
}
