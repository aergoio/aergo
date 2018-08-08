/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"reflect"
	"time"

	"github.com/aergoio/aergo-actor/actor"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logger = log.NewLogger(log.RPC)
)

// AergoRPCService implements GRPC server which is defined in rpc.proto
type AergoRPCService struct {
	hub         *component.ComponentHub
	actorHelper p2p.ActorService
	msgHelper   message.Helper
}

// FIXME remove redundent constants
const halfMinute = time.Second * 30
const defaultActorTimeout = time.Second * 3

var _ types.AergoRPCServiceServer = (*AergoRPCService)(nil)

// Blockchain handle rpc request blockchain. It has no additional input parameter
func (rpc *AergoRPCService) Blockchain(ctx context.Context, in *types.Empty) (*types.BlockchainStatus, error) {
	//last, _ := rpc.ChainService.GetBestBlock()
	result, err := rpc.hub.RequestFuture(message.ChainSvc, &message.GetBestBlock{}, defaultActorTimeout,
		"rpc.(*AergoRPCService).Blockchain").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetBestBlockRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type error")
	}
	if rsp.Err != nil {
		return nil, rsp.Err
	}
	last := rsp.Block
	return &types.BlockchainStatus{
		BestBlockHash: last.GetHash(),
		BestHeight:    last.GetHeader().GetBlockNo(),
	}, nil
}

// ListBlockHeaders handle rpc request listblocks
func (rpc *AergoRPCService) ListBlockHeaders(ctx context.Context, in *types.ListParams) (*types.BlockHeaderList, error) {
	var maxFetchSize uint32
	// TODO refactor with almost same code is in p2pcmdblock.go
	if in.Size > uint32(1000) {
		maxFetchSize = uint32(1000)
	} else {
		maxFetchSize = in.Size
	}
	idx := uint32(0)
	hashes := make([][]byte, 0, maxFetchSize)
	headers := make([]*types.Block, 0, maxFetchSize)
	if len(in.Hash) > 0 {
		hash := in.Hash
		for idx < maxFetchSize {
			foundBlock, ok := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
				&message.GetBlock{BlockHash: hash}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#1"))
			if !ok || nil == foundBlock {
				break
			}
			hashes = append(hashes, foundBlock.Hash)
			foundBlock.Body = nil
			headers = append(headers, foundBlock)
			idx++
			hash = foundBlock.Header.PrevBlockHash
			if len(hash) == 0 {
				break
			}
		}
	} else {
		end := types.BlockNo(0)
		if types.BlockNo(in.Height) >= types.BlockNo(maxFetchSize) {
			end = types.BlockNo(in.Height) - types.BlockNo(maxFetchSize-1)
		}
		for i := types.BlockNo(in.Height); i >= end; i-- {
			foundBlock, ok := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
				&message.GetBlockByNo{BlockNo: i}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#2"))
			if !ok || nil == foundBlock {
				break
			}
			hashes = append(hashes, foundBlock.Hash)
			foundBlock.Body = nil
			headers = append(headers, foundBlock)
			idx++
		}
	}

	return &types.BlockHeaderList{Blocks: headers}, nil

}

func extractBlockFromFuture(future *actor.Future) (*types.Block, bool) {
	rawResponse, err := future.Result()
	if err != nil {
		return nil, false
	}
	var blockRsp *message.GetBlockRsp
	switch v := rawResponse.(type) {
	case message.GetBlockRsp:
		blockRsp = &v
	case message.GetBestBlockRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	case message.GetBlockByNoRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	default:
		return nil, false
	}
	return extractBlock(blockRsp)
}

func extractBlock(from *message.GetBlockRsp) (*types.Block, bool) {
	if nil != from.Err {
		return nil, false
	}
	return from.Block, true

}

// GetBlock handle rpc request getblock
func (rpc *AergoRPCService) GetBlock(ctx context.Context, in *types.SingleBytes) (*types.Block, error) {
	var result interface{}
	var err error
	if cap(in.Value) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "recevice no bytes")
	}
	if len(in.Value) < 32 {
		number := uint64(binary.LittleEndian.Uint64(in.Value))
		result, err = rpc.hub.RequestFuture(message.ChainSvc, &message.GetBlockByNo{BlockNo: number},
			defaultActorTimeout, "rpc.(*AergoRPCService).GetBlock#1").Result()
	} else {
		result, err = rpc.hub.RequestFuture(message.ChainSvc, &message.GetBlock{BlockHash: in.Value},
			defaultActorTimeout, "rpc.(*AergoRPCService).GetBlock#2").Result()
	}
	if err != nil {
		return nil, err
	}
	found, err := rpc.msgHelper.ExtractBlockFromResponse(result)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "not found")
	}
	return found, nil
}

// GetTX handle rpc request gettx
func (rpc *AergoRPCService) GetTX(ctx context.Context, in *types.SingleBytes) (*types.Tx, error) {
	result, err := rpc.actorHelper.CallRequest(message.MemPoolSvc,
		&message.MemPoolExist{Hash: in.Value})
	tx, err := rpc.msgHelper.ExtractTxFromResponse(result)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		return tx, nil
	}
	// TODO try find tx in blockchain, but chainservice doesn't have method yet.

	return nil, status.Errorf(codes.NotFound, "not found")
}

// GetBlockTX handle rpc request gettx
func (rpc *AergoRPCService) GetBlockTX(ctx context.Context, in *types.SingleBytes) (*types.TxInBlock, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetTx{TxHash: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetBlockTX").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return &types.TxInBlock{Tx: rsp.Tx, TxIdx: rsp.TxIds}, rsp.Err
}

var emptyBytes = make([]byte, 0)

// CommitTX handle rpc request commit
func (rpc *AergoRPCService) CommitTX(ctx context.Context, in *types.TxList) (*types.CommitResultList, error) {
	// TODO: check validity
	//if bytes.Equal(emptyBytes, in.Hash) {
	//	return nil, status.Errorf(codes.InvalidArgument, "invalid hash")
	//}
	if in.Txs == nil {
		return nil, status.Errorf(codes.InvalidArgument, "input tx is empty")
	}
	rs := make([]*types.CommitResult, len(in.Txs))
	results := &types.CommitResultList{Results: rs}
	//results := &types.CommitResultList{}
	start := 0
	cnt := 0
	chunk := 100

	for i, tx := range in.Txs {
		hash := tx.Hash
		var r types.CommitResult
		r.Hash = hash

		calculated := tx.CalculateTxHash()

		if !bytes.Equal(hash, calculated) {
			r.Error = types.CommitStatus_COMMIT_STATUS_INVALID_ARGUMENT
		}
		results.Results[i] = &r
		cnt++

		if (i > 0 && i%chunk == 0) || i == len(in.Txs)-1 {
			//send tx message to mempool
			result, err := rpc.hub.RequestFuture(message.MemPoolSvc,
				&message.MemPoolPut{Txs: in.Txs[start : start+cnt]},
				defaultActorTimeout, "rpc.(*AergoRPCService).CommitTX").Result()
			if err != nil {
				return nil, err
			}
			rsp, ok := result.(*message.MemPoolPutRsp)
			if !ok {
				return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
			}

			for j, err := range rsp.Err {
				switch err {
				case nil:
					results.Results[start+j].Error = types.CommitStatus_COMMIT_STATUS_OK
				case message.ErrTxNonceTooLow:
					results.Results[start+j].Error = types.CommitStatus_COMMIT_STATUS_NONCE_TOO_LOW
				case message.ErrTxAlreadyInMempool:
					results.Results[start+j].Error = types.CommitStatus_COMMIT_STATUS_TX_ALREADY_EXISTS
				default:
					results.Results[start+j].Error = types.CommitStatus_COMMIT_STATUS_TX_INTERNAL_ERROR

				}
			}
			start += cnt
			cnt = 0
		}

	}

	return results, nil
}

// GetState handle rpc request getstate
func (rpc *AergoRPCService) GetState(ctx context.Context, in *types.SingleBytes) (*types.State, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetState{Account: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetState").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetStateRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	//TODO : rsp.Account will be filled in ChainSvc?
	rsp.State.Account = in.Value
	return rsp.State, rsp.Err
}

// CreateAccount handle rpc request newaccount
func (rpc *AergoRPCService) CreateAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.CreateAccount{Passphrase: in.Passphrase}, defaultActorTimeout, "rpc.(*AergoRPCService).CreateAccount").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.CreateAccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, nil
}

// GetAccounts handle rpc request getaccounts
func (rpc *AergoRPCService) GetAccounts(ctx context.Context, in *types.Empty) (*types.AccountList, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.GetAccounts{}, defaultActorTimeout, "rpc.(*AergoRPCService).GetAccounts").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetAccountsRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Accounts, nil
}

// LockAccount handle rpc request lockaccount
func (rpc *AergoRPCService) LockAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.LockAccount{Account: in.Account, Passphrase: in.Passphrase},
		defaultActorTimeout, "rpc.(*AergoRPCService).LockAccount").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.AccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, rsp.Err
}

// UnlockAccount handle rpc request unlockaccount
func (rpc *AergoRPCService) UnlockAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.UnlockAccount{Account: in.Account, Passphrase: in.Passphrase},
		defaultActorTimeout, "rpc.(*AergoRPCService).UnlockAccount").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.AccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, rsp.Err
}

// SignTX handle rpc request signtx
func (rpc *AergoRPCService) SignTX(ctx context.Context, in *types.Tx) (*types.Tx, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.SignTx{Tx: in}, defaultActorTimeout, "rpc.(*AergoRPCService).SignTX").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.SignTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Tx, rsp.Err
}

// VerifyTX handle rpc request verifytx
func (rpc *AergoRPCService) VerifyTX(ctx context.Context, in *types.Tx) (*types.VerifyResult, error) {
	result, err := rpc.hub.RequestFuture(message.AccountsSvc,
		&message.VerifyTx{Tx: in}, defaultActorTimeout, "rpc.(*AergoRPCService).VerifyTX").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.VerifyTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	ret := &types.VerifyResult{Tx: rsp.Tx}
	if rsp.Err == message.ErrSignNotMatch {
		ret.Error = types.VerifyStatus_VERIFY_STATUS_SIGN_NOT_MATCH
	} else {
		ret.Error = types.VerifyStatus_VERIFY_STATUS_OK
	}
	return ret, nil
}

// GetPeers handle rpc request getpeers
func (rpc *AergoRPCService) GetPeers(ctx context.Context, in *types.Empty) (*types.PeerList, error) {
	result, err := rpc.hub.RequestFuture(message.P2PSvc,
		&message.GetPeers{}, halfMinute, "rpc.(*AergoRPCService).GetPeers").Result()
	if err != nil {
		return nil, err
	}
	rsp := result.(*message.GetPeersRsp)
	states := make([]int32, len(rsp.States))
	for i, state := range rsp.States {
		states[i] = int32(state)
	}

	return &types.PeerList{Peers: rsp.Peers, States: states}, nil
}

// State handle rpc request state
func (rpc *AergoRPCService) NodeState(ctx context.Context, in *types.Empty) (*types.NodeStatus, error) {
	//result, err := rpc.hub.RequestFuture(message.P2PSvc,
	status := rpc.hub.Status()

	result := &types.NodeStatus{}
	for k, v := range status {
		module := &types.ModuleStatus{
			Name: k,
		}
		for ik, iv := range v.GetAll() {
			for iik, iiv := range iv {
				var stat float64
				switch value := iiv.(type) {
				case int64:
					stat = float64(value)
				case float64:
					stat = value
				default:
					logger.Warnf("unresolve value in  node state: %v", value)
				}
				internal := &types.InternalStat{
					Name: ik + "/" + iik,
					Stat: stat,
				}
				module.Stat = append(module.Stat, internal)
			}
		}
		result.Status = append(result.Status, module)
	}

	return result, nil
}

func toTimestamp(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: time.Unix(),
		Nanos:   int32(time.Nanosecond())}
}

func fromTimestamp(timestamp *timestamp.Timestamp) time.Time {
	return time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
}
