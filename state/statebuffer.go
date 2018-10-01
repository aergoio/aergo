package state

import (
	"sort"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var (
	emptyBufferEntry = bufferEntry{}
)

type bufferEntry struct {
	key  types.HashID
	hash types.HashID
	data interface{}
}

func newBufferEntry(key types.HashID, hash types.HashID, data interface{}) *bufferEntry {
	return &bufferEntry{
		key:  key,
		hash: hash,
		data: data,
	}
}

func (et *bufferEntry) getKey() []byte {
	return et.key[:]
}
func (et *bufferEntry) getHash() []byte {
	return et.hash[:]
}
func (et *bufferEntry) getData() interface{} {
	return et.data
}

type bufferIndex map[types.HashID]int

type stateBuffer struct {
	entries []bufferEntry
	indexes bufferIndex
}

func newStateBuffer() *stateBuffer {
	buffer := stateBuffer{
		entries: []bufferEntry{},
		indexes: bufferIndex{},
	}
	return &buffer
}

func (buffer *stateBuffer) reset() error {
	// TODO
	buffer.entries = buffer.entries[:0]
	buffer.indexes = bufferIndex{}
	return nil
}

func (buffer *stateBuffer) marshal(data interface{}) ([]byte, error) {
	// TODO
	switch data.(type) {
	case ([]byte):
		return data.([]byte), nil
	case (*[]byte):
		return *(data.(*[]byte)), nil
	case (proto.Message):
		return proto.Marshal(data.(proto.Message))
	}
	return nil, nil
}

func (buffer *stateBuffer) get(key types.HashID) *bufferEntry {
	if index, ok := buffer.indexes[key]; ok {
		return &buffer.entries[index]
	}
	return nil
}

// func (buffer *stateBuffer) puts(ets ...bufferEntry) {
// 	for _, v := range ets {
// 		buffer.entries = append(buffer.entries, v)
// 		buffer.indexes[v.key] = buffer.snapshot()
// 	}
// }

func (buffer *stateBuffer) put(key types.HashID, data interface{}) error {
	buf, err := buffer.marshal(data)
	if err != nil {
		return err
	}
	hash := types.GetHashID(buf)
	et := newBufferEntry(key, hash, data)
	buffer.entries = append(buffer.entries, *et)
	buffer.indexes[key] = buffer.snapshot()
	return nil
}

func (buffer *stateBuffer) snapshot() int {
	// TODO: last index of entries
	return len(buffer.entries) - 1
}

func (buffer *stateBuffer) rollback(snapshot int) error {
	// TODO: rollback entries and indexes
	return nil
}

func (buffer *stateBuffer) isEmpty() bool {
	return len(buffer.entries) == 0
}

func (buffer *stateBuffer) export() ([][]byte, [][]byte) {
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
		keys[i] = append(keys[i], et.getKey()...)
		vals[i] = append(vals[i], et.getHash()...)
	}
	return keys, vals
}

func (buffer *stateBuffer) commit(store *db.DB) error {
	dbtx := (*store).NewTx()
	for _, v := range buffer.indexes {
		et := buffer.entries[v]
		buf, err := buffer.marshal(et.data)
		if err != nil {
			dbtx.Discard()
			return err
		}
		dbtx.Set(et.getHash(), buf)
	}
	dbtx.Commit()
	return nil
}
