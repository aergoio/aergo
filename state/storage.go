package state

import (
	"bytes"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
)

var (
	checkpointKey = types.ToHashID([]byte("checkpoint"))
)

type storageCache map[types.AccountID]*bufferedStorage

func newStorageCache() *storageCache {
	return &storageCache{}
}

func (cache *storageCache) get(key types.AccountID) *bufferedStorage {
	if storage, ok := (*cache)[key]; ok && storage != nil {
		return storage
	}
	return nil
}
func (cache *storageCache) put(key types.AccountID, storage *bufferedStorage) {
	(*cache)[key] = storage
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

func (storage *bufferedStorage) stage(dbtx *db.Transaction) error {
	storage.trie.StageUpdates(dbtx)
	if err := storage.buffer.stage(dbtx); err != nil {
		return err
	}
	if err := storage.buffer.reset(); err != nil {
		return err
	}
	return nil
}
