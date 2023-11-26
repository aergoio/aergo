package state

import (
	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/internal/enc/gob"
	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/types"
)

func saveData(store db.DB, key []byte, data interface{}) error {
	if key == nil {
		return errSaveData
	}
	var err error
	var raw []byte
	switch msg := data.(type) {
	case ([]byte):
		raw = msg
	case proto.Message:
		raw, err = proto.Encode(msg)
		if err != nil {
			return err
		}
	default:
		raw, err = gob.Encode(msg)
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
	switch msg := data.(type) {
	case *[]byte:
		*msg = raw
	case proto.Message:
		err = proto.Decode(raw, msg)
	default:
		err = gob.Decode(raw, msg)
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
