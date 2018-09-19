package state

import (
	"bytes"
	"encoding/gob"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

func saveData(store *db.DB, key []byte, data interface{}) error {
	if key == nil {
		return errSaveData
	}
	var err error
	var raw []byte
	switch data.(type) {
	case (*[]byte):
		raw = *(data.(*[]byte))
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
	(*store).Set(key, raw)
	return nil
}

func loadData(store *db.DB, key []byte, data interface{}) error {
	if key == nil {
		return errLoadData
	}
	if !(*store).Exist(key) {
		return nil
	}
	raw := (*store).Get(key)

	if raw == nil || len(raw) == 0 {
		return nil
	}
	var err error
	switch data.(type) {
	case (*[]byte):
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

func (sdb *ChainStateDB) saveStateDB() error {
	err := saveData(sdb.statedb, []byte(stateLatest), sdb.latest)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *ChainStateDB) loadStateDB() error {
	err := loadData(sdb.statedb, []byte(stateLatest), &sdb.latest)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *ChainStateDB) saveBlockState(data *types.BlockState) error {
	bid := data.BlockHash
	if bid == emptyBlockID {
		return errSaveBlockState
	}
	err := saveData(sdb.statedb, bid[:], data)
	return err
}
func (sdb *ChainStateDB) loadBlockState(bid types.BlockID) (*types.BlockState, error) {
	if bid == emptyBlockID {
		return nil, errLoadBlockState
	}
	data := &types.BlockState{}
	err := loadData(sdb.statedb, bid[:], data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (sdb *ChainStateDB) loadStateData(key []byte) (*types.State, error) {
	if key == nil || len(key) == 0 {
		return nil, errLoadStateData
	}
	data := &types.State{}
	err := loadData(sdb.statedb, key, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
