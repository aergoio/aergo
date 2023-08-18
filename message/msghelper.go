package message

import (
	"fmt"
	"reflect"

	"github.com/aergoio/aergo/v2/types"
)

const RPCSvc = "RPCSvc"

// Helper is helper interface for extracting tx or block from actor response
type Helper interface {
	// ExtractBlockFromResponseAndError get rawResponse and error and return pointer of Block
	ExtractBlockFromResponseAndError(rawResponse interface{}, err error) (*types.Block, error)
	ExtractBlockFromResponse(rawResponse interface{}) (*types.Block, error)
	ExtractTxFromResponseAndError(rawResponse interface{}, err error) (*types.Tx, error)
	ExtractTxFromResponse(rawResponse interface{}) (*types.Tx, error)
	ExtractTxsFromResponseAndError(rawResponse interface{}, err error) ([]*types.Tx, error)
	ExtractTxsFromResponse(rawResponse interface{}) ([]*types.Tx, error)
}

func GetHelper() Helper {
	return &baseHelper{}
}

type baseHelper struct {
}

func (h baseHelper) ExtractBlockFromResponseAndError(rawResponse interface{}, err error) (*types.Block, error) {
	if err != nil {
		return nil, err
	}
	return h.ExtractBlockFromResponse(rawResponse)
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

func (h baseHelper) ExtractTxFromResponseAndError(rawResponse interface{}, err error) (*types.Tx, error) {
	if err != nil {
		return nil, err
	}
	return h.ExtractTxFromResponse(rawResponse)
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

func (h baseHelper) ExtractTxsFromResponseAndError(rawResponse interface{}, err error) ([]*types.Tx, error) {
	if err != nil {
		return nil, err
	}
	return h.ExtractTxsFromResponse(rawResponse)
}

func (h baseHelper) ExtractTxsFromResponse(rawResponse interface{}) ([]*types.Tx, error) {
	switch v := rawResponse.(type) {
	case *MemPoolGetRsp:
		if v.Err != nil {
			return nil, v.Err
		}
		res := make([]*types.Tx, 0)
		for _, x := range v.Txs {
			res = append(res, x.GetTx())
		}
		return res, nil
	case *MemPoolExistExRsp:
		return v.Txs, nil
	default:
		panic(fmt.Sprintf("unexpected result type %s, expected %s", reflect.TypeOf(rawResponse),
			"message.MemPoolGetRsp"))
	}
}
