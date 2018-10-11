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

	if len(raw) == 0 {
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

func (sdb *ChainStateDB) saveStateLatest() error {
	return saveData(&sdb.store, []byte(stateLatest), sdb.latest)
}

func (sdb *ChainStateDB) loadStateLatest() error {
	return loadData(&sdb.store, []byte(stateLatest), &sdb.latest)
}

func (sdb *ChainStateDB) saveBlockInfo(data *BlockInfo) error {
	bid := data.BlockHash
	if bid == emptyBlockID {
		return errSaveBlockInfo
	}

	err := saveData(&sdb.store, bid[:], data)
	return err
}
func (sdb *ChainStateDB) loadBlockInfo(bid types.BlockID) (*BlockInfo, error) {
	if bid == emptyBlockID {
		return nil, errLoadBlockInfo
	}
	data := &BlockInfo{}
	err := loadData(&sdb.store, bid[:], data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (states *StateDB) loadStateData(key []byte) (*types.State, error) {
	if len(key) == 0 {
		return nil, errLoadStateData
	}
	data := &types.State{}
	err := loadData(states.store, key, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
