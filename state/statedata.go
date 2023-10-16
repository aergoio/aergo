package state

import (
	"bytes"
	"encoding/gob"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
)

func saveData(store db.DB, key []byte, data interface{}) error {
	if key == nil {
		return errSaveData
	}
	var err error
	var raw []byte
	switch data.(type) {
	case ([]byte):
		raw = data.([]byte)
	case proto.Message:
		raw, err = proto.Marshal(data.(proto.Message))
		if err != nil {
			return err
		}
	default:
		buffer := &bytes.Buffer{}
		enc := gob.NewEncoder(buffer)
		err = enc.Encode(data)
		if err != nil {
			return err
		}
		raw = buffer.Bytes()
		if err != nil {
			return err
		}
	}
	store.Set(key, raw)
	return nil
}

func loadData(store db.DB, key []byte, data interface{}) error {
	if key == nil {
		return errLoadData
	}
	raw := store.Get(key)

	if len(raw) == 0 {
		return nil
	}
	var err error
	switch data.(type) {
	case *[]byte:
		*(data).(*[]byte) = raw
	case proto.Message:
		err = proto.Unmarshal(raw, data.(proto.Message))
	default:
		reader := bytes.NewReader(raw)
		dec := gob.NewDecoder(reader)
		err = dec.Decode(data)
	}
	return err
}

func (states *StateDB) loadStateData(key []byte) (*types.State, error) {
	if len(key) == 0 {
		return nil, errLoadStateData
	}
	data := &types.State{}
	if err := loadData(states.store, key, data); err != nil {
		return nil, err
	}
	return data, nil
}
