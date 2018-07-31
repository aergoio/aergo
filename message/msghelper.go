package message

import (
	"fmt"
	"reflect"

	"github.com/aergoio/aergo/types"
)

// Helper is helper interface for extracting tx or block from actor response
type Helper interface {
	ExtractBlockFromResponse(rawResponse interface{}) (*types.Block, error)
	ExtractTxFromResponse(rawResponse interface{}) (*types.Tx, error)
	ExtractTxsFromResponse(rawResponse interface{}) ([]*types.Tx, error)
}

func GetHelper() Helper {
	return &baseHelper{}
}

type baseHelper struct {
}

func (h baseHelper) ExtractBlockFromResponse(rawResponse interface{}) (*types.Block, error) {
	var blockRsp *GetBlockRsp
	switch v := rawResponse.(type) {
	case GetBlockRsp:
		blockRsp = &v
	case GetBestBlockRsp:
		blockRsp = (*GetBlockRsp)(&v)
	case GetBlockByNoRsp:
		blockRsp = (*GetBlockRsp)(&v)
	default:
		panic(fmt.Sprintf("unexpected result type %s, expected %s", reflect.TypeOf(rawResponse),
			"message.GetBlockRsp"))
	}
	return extractBlock(blockRsp)
}

func extractBlock(from *GetBlockRsp) (*types.Block, error) {
	if nil != from.Err {
		return nil, from.Err
	}
	return from.Block, nil
}

func (h baseHelper) ExtractTxFromResponse(rawResponse interface{}) (*types.Tx, error) {
	switch v := rawResponse.(type) {
	case *MemPoolExistRsp:
		return v.Tx, nil
	default:
		panic(fmt.Sprintf("unexpected result type %s, expected %s", reflect.TypeOf(rawResponse),
			"message.MemPoolExistRsp"))
	}
}

func (h baseHelper) ExtractTxsFromResponse(rawResponse interface{}) ([]*types.Tx, error) {
	switch v := rawResponse.(type) {
	case *MemPoolGetRsp:
		if v.Err != nil {
			return nil, v.Err
		}
		return v.Txs, nil
	default:
		panic(fmt.Sprintf("unexpected result type %s, expected %s", reflect.TypeOf(rawResponse),
			"message.MemPoolGetRsp"))
	}
}
