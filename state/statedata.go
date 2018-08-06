package state

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

func (sdb *CachedStateDB) saveData(key []byte, data interface{}) error {
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
	}
	logger.Debugf("- putData: key=%v, size=%d", hex.EncodeToString(key), len(raw))
	sdb.statedb.Set(key, raw)
	return nil
}

func (sdb *CachedStateDB) loadData(key []byte, data interface{}) error {
	if key == nil {
		return fmt.Errorf("Failed to get data: key is nil")
	}
	if !sdb.statedb.Exist(key) {
		return nil
	}
	raw := sdb.statedb.Get(key)
	logger.Debugf("- loadData: key=%v, size=%d", hex.EncodeToString(key), len(raw))
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
	return err
}

func (sdb *CachedStateDB) saveStateDB() error {
	logger.Debug("- sdb.latest.blockNo=", sdb.latest.blockNo)
	logger.Debug("- sdb.latest.blockHash=", sdb.latest.blockHash)
	logger.Debug("- sdb.accounts.size=", len(sdb.accounts))
	err := sdb.saveData([]byte(stateAccounts), sdb.accounts)
	if err != nil {
		return err
	}
	err = sdb.saveData([]byte(stateLatest), sdb.latest)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *CachedStateDB) loadStateDB() error {
	err := sdb.loadData([]byte(stateLatest), sdb.latest)
	if err != nil {
		return err
	}
	err = sdb.loadData([]byte(stateAccounts), sdb.accounts)
	if err != nil {
		return err
	}
	logger.Debug("- sdb.latest.blockNo=", sdb.latest.blockNo)
	logger.Debug("- sdb.latest.blockHash=", sdb.latest.blockHash)
	logger.Debug("- sdb.accounts.size=", len(sdb.accounts))
	return nil
}

func (sdb *CachedStateDB) saveBlockState(data *BlockState) error {
	bkey := data.blockHash
	if bkey == types.EmptyBlockKey {
		return fmt.Errorf("Invalid Key to save BlockState: empty")
	}
	err := sdb.saveData(bkey[:], data)
	return err
}
func (sdb *CachedStateDB) loadBlockState(bkey types.BlockKey) (*BlockState, error) {
	if bkey == types.EmptyBlockKey {
		return nil, fmt.Errorf("Invalid Key to load BlockState: empty")
	}
	data := &BlockState{}
	err := sdb.loadData(bkey[:], data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
