package state

import (
	"sort"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

type entry interface {
	KeyID() types.HashID
	HashID() types.HashID
	Value() interface{}
}

type cached interface {
	cache() *stateBuffer
}

type valueEntry struct {
	key   types.HashID
	hash  types.HashID
	value interface{}
}

func newValueEntry(key types.HashID, value interface{}) entry {
	return &valueEntry{
		key:   key,
		hash:  getHash(value),
		value: value,
	}
}
func (et *valueEntry) KeyID() types.HashID {
	return et.key
}
func (et *valueEntry) HashID() types.HashID {
	return et.hash
}
func (et *valueEntry) Value() interface{} {
	return et.value
}

type metaEntry struct {
	*valueEntry
}

func newMetaEntry(key types.HashID, value interface{}) entry {
	return &metaEntry{
		valueEntry: &valueEntry{
			key:   key,
			value: value,
		},
	}
}

type bufferIndex map[types.HashID]*stack

func (idxs *bufferIndex) peek(key types.HashID) int {
	return (*idxs)[key].peek()
}
func (idxs *bufferIndex) pop(key types.HashID) int {
	return (*idxs)[key].pop()
}
func (idxs *bufferIndex) push(key types.HashID, argv ...int) {
	(*idxs)[key] = (*idxs)[key].push(argv...)
}
func (idxs *bufferIndex) rollback(snapshot int) {
	for k, v := range *idxs {
		for v.peek() >= snapshot {
			v.pop()
		}
		if v.peek() < 0 {
			delete(*idxs, k)
		}
	}
}

type stateBuffer struct {
	entries []entry
	indexes bufferIndex
	nextIdx int
}

func newStateBuffer() *stateBuffer {
	buffer := stateBuffer{
		entries: []entry{},
		indexes: bufferIndex{},
		nextIdx: 0,
	}
	return &buffer
}

func (buffer *stateBuffer) reset() error {
	return buffer.rollback(0)
}

func (buffer *stateBuffer) get(key types.HashID) entry {
	if index, ok := buffer.indexes[key]; ok {
		return buffer.entries[index.peek()]
	}
	return nil
}

func (buffer *stateBuffer) lookup(key types.HashID, match func(et entry) bool) entry {
	it := buffer.indexes[key].iter()
	for rev := it(); rev >= 0; rev = it() {
		et := buffer.entries[rev]
		if match(et) {
			return et
		}
	}
	return nil
}

func (buffer *stateBuffer) lookupCache(key types.HashID) *stateBuffer {
	res := buffer.lookup(key, func(et entry) bool {
		switch et.Value().(type) {
		case cached:
			return true
		}
		return false
	})
	if res != nil {
		return res.Value().(cached).cache()
	}
	return nil
}

func (buffer *stateBuffer) put(et entry) {
	snapshot := buffer.snapshot()
	buffer.entries = append(buffer.entries, et)
	buffer.indexes[et.KeyID()] = buffer.indexes[et.KeyID()].push(snapshot)
	buffer.nextIdx++
}

func (buffer *stateBuffer) snapshot() int {
	return buffer.nextIdx
}

func (buffer *stateBuffer) rollback(snapshot int) error {
	for i := buffer.nextIdx - 1; i >= snapshot; i-- {
		et := buffer.entries[i]
		buffer.indexes.pop(et.KeyID())
		idx := buffer.indexes.peek(et.KeyID())
		if idx < 0 {
			delete(buffer.indexes, et.KeyID())
			continue
		}
	}
	buffer.entries = buffer.entries[:snapshot]
	//buffer.indexes.rollback(snapshot)
	buffer.nextIdx = snapshot
	return nil
}

func (buffer *stateBuffer) isEmpty() bool {
	return len(buffer.entries) == 0
}

func (buffer *stateBuffer) export() ([][]byte, [][]byte) {
	bufs := make([]entry, 0, len(buffer.indexes))
	for _, v := range buffer.indexes {
		idx := v.peek()
		if idx < 0 {
			continue
		}
		et := buffer.entries[idx]
		if _, ok := et.(metaEntry); ok {
			// skip meta entry
			continue
		}
		bufs = append(bufs, et)
	}
	sort.Slice(bufs, func(i, j int) bool {
		return -1 == (bufs[i].KeyID().Compare(bufs[j].KeyID()))
	})
	size := len(bufs)
	keys := make([][]byte, size)
	vals := make([][]byte, size)
	for i, et := range bufs {
		keys[i] = append(keys[i], et.KeyID().Bytes()...)
		vals[i] = append(vals[i], et.HashID().Bytes()...)
	}
	return keys, vals
}

func (buffer *stateBuffer) updateTrie(tr *trie.Trie) error {
	keys, vals := buffer.export()
	if len(keys) == 0 || len(vals) == 0 {
		// nothing to update
		return nil
	}
	if _, err := tr.Update(keys, vals); err != nil {
		return err
	}
	return nil
}

// func (buffer *stateBuffer) stageUpdates(tr *trie.Trie, batchtx *db.Transaction) error {
// 	// keys, vals := buffer.export()
// 	// if len(keys) == 0 || len(vals) == 0 {
// 	// 	// nothing to update
// 	// 	return nil
// 	// }
// 	// if _, err := tr.Update(keys, vals); err != nil {
// 	// 	return err
// 	// }
// 	if err := buffer.updateTrie(tr); err != nil {
// 		return err
// 	}
// 	tr.StageUpdates(batchtx)
// 	if err := buffer.stage(batchtx); err != nil {
// 		return err
// 	}
// 	if err := buffer.reset(); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (buffer *stateBuffer) stage(dbtx *db.Transaction) error {
	for _, v := range buffer.indexes {
		et := buffer.entries[v.peek()]
		buf, err := marshal(et.Value())
		if err != nil {
			return err
		}
		(*dbtx).Set(et.HashID().Bytes(), buf)
	}
	return nil
}

func marshal(data interface{}) ([]byte, error) {
	switch data.(type) {
	case ([]byte):
		return data.([]byte), nil
	case (*[]byte):
		return *(data.(*[]byte)), nil
	case (types.ImplMarshal):
		return data.(types.ImplMarshal).Marshal()
	case (proto.Message):
		return proto.Marshal(data.(proto.Message))
	}
	return nil, nil
}

func getHash(data interface{}) types.HashID {
	switch data.(type) {
	case (types.ImplHashID):
		return data.(types.ImplHashID).HashID()
	default:
	}
	buf, err := marshal(data)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get hash: marshal")
		return emptyHashID
	}
	return types.GetHashID(buf)
}
