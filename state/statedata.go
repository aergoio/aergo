package state

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

func (sdb *CachedStateDB) GetState(skey types.StateKey) (*types.State, error) {
	if state, ok := sdb.cache[skey]; ok {
		return state, nil
	}
	buf := &types.State{}
	err := sdb.getData(skey[:], buf)
	if err != nil {
		return nil, err
	}
	state := buf
	sdb.cache[skey] = state
	return state, nil
}
func (sdb *CachedStateDB) PutState(skey types.StateKey, state *types.State) error {
	sdb.cache[skey] = state
	err := sdb.putData(skey[:], state)
	if err != nil {
		return err
	}
	return nil
}

func (sdb *CachedStateDB) putData(key []byte, data interface{}) error {
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
	// logger.Debugf("putData: key=%v, raw(%d)=%v",
	// 	hex.EncodeToString(key), len(raw), hex.EncodeToString(raw))
	sdb.statedb.Set(key, raw)
	return nil
}

func (sdb *CachedStateDB) getData(key []byte, data interface{}) error {
	if key == nil {
		return fmt.Errorf("Failed to get data: key is nil")
	}
	if !sdb.statedb.Exist(key) {
		return nil
	}
	raw := sdb.statedb.Get(key)
	// logger.Debugf("getData: key=%v, raw(%d)=%v",
	// 	hex.EncodeToString(key), len(raw), hex.EncodeToString(raw))
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
