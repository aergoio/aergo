package state

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

func saveData(store *db.DB, key []byte, data interface{}) error {
	if key == nil {
		return fmt.Errorf("Failed to set data: key is nil")
	}
	var err error
	var raw []byte
	switch data.(type) {
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
	// logger.Debugf("- saveData: key=%v, size=%d", hex.EncodeToString(key), len(raw))
	// logger.Debugf("- saveData: data=%v", data)
	(*store).Set(key, raw)
	return nil
}

func loadData(store *db.DB, key []byte, data interface{}) error {
	if key == nil {
		return fmt.Errorf("Failed to get data: key is nil")
	}
	if !(*store).Exist(key) {
		return nil
	}
	raw := (*store).Get(key)

	// logger.Debugf("- loadData: key=%v, size=%d", hex.EncodeToString(key), len(raw))
	if raw == nil || len(raw) == 0 {
		return nil
	}
	var err error
	switch data.(type) {
	case proto.Message:
		err = proto.Unmarshal(raw, data.(proto.Message))
	default:
		reader := bytes.NewReader(raw)
		dec := gob.NewDecoder(reader)
		err = dec.Decode(data)
	}
	// logger.Debugf("- loadData: data=%v", data)
	return err
}

func (sdb *ChainStateDB) saveStateDB() error {
	// logger.Debugf("- ### saveStateDB")
	// logger.Debugf("- sdb.latest: BlockNo=%d, BlockHash=%s", sdb.latest.BlockNo, sdb.latest.BlockHash)
	// logger.Debugf("- sdb.accounts: size=%d", len(sdb.accounts))
	err := saveData(sdb.statedb, []byte(stateAccounts), sdb.accounts)
	if err != nil {
		return err
	}
	err = saveData(sdb.statedb, []byte(stateLatest), sdb.latest)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *ChainStateDB) loadStateDB() error {
	// logger.Debug("- ### loadStateDB")
	err := loadData(sdb.statedb, []byte(stateLatest), &sdb.latest)
	if err != nil {
		return err
	}
	// logger.Debugf("- sdb.latest: BlockNo=%d, BlockHash=%s", sdb.latest.BlockNo, sdb.latest.BlockHash)
	err = loadData(sdb.statedb, []byte(stateAccounts), &sdb.accounts)
	if err != nil {
		return err
	}
	// logger.Debugf("- sdb.accounts: size=%d", len(sdb.accounts))
	return nil
}

func (sdb *ChainStateDB) saveBlockState(data *BlockState) error {
	bid := data.BlockHash
	if bid == emptyBlockID {
		return fmt.Errorf("Invalid ID to save BlockState: empty")
	}
	err := saveData(sdb.statedb, bid[:], data)
	return err
}
func (sdb *ChainStateDB) loadBlockState(bid types.BlockID) (*BlockState, error) {
	if bid == emptyBlockID {
		return nil, fmt.Errorf("Invalid ID to load BlockState: empty")
	}
	data := &BlockState{}
	err := loadData(sdb.statedb, bid[:], data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
