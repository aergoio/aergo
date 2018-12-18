/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p"
	"github.com/aergoio/aergo/p2p/metric"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logger = log.NewLogger("rpc")
)

// AergoRPCService implements GRPC server which is defined in rpc.proto
type AergoRPCService struct {
	hub         *component.ComponentHub
	actorHelper p2p.ActorService
	msgHelper   message.Helper

	blockStreamLock         sync.RWMutex
	blockStream             []types.AergoRPCService_ListBlockStreamServer
	blockMetadataStreamLock sync.RWMutex
	blockMetadataStream     []types.AergoRPCService_ListBlockMetadataStreamServer
}

// FIXME remove redundant constants
const halfMinute = time.Second * 30
const defaultActorTimeout = time.Second * 3

var _ types.AergoRPCServiceServer = (*AergoRPCService)(nil)

func (rpc *AergoRPCService) Metric(ctx context.Context, req *types.MetricsRequest) (*types.Metrics, error) {
	result := &types.Metrics{}
	processed := make(map[types.MetricType]interface{})
	for _, mt := range req.Types {
		if _, found := processed[mt]; found {
			continue
		}
		processed[mt] = mt

		switch mt {
		case types.MetricType_P2P_NETWORK:
			rpc.fillPeerMetrics(result)
		default:
			// TODO log itB
		}
	}

	return result, nil
}

func (rpc *AergoRPCService) fillPeerMetrics(result *types.Metrics) {
	// fill metrics for p2p
	presult, err := rpc.actorHelper.CallRequestDefaultTimeout(message.P2PSvc,
		&message.GetMetrics{})
	if err != nil {
		return
	}
	metrics := presult.([]*metric.PeerMetric)
	mets := make([]*types.PeerMetric, len(metrics))
	for i, met := range metrics {
		rMet := &types.PeerMetric{PeerID: []byte(met.PeerID), SumIn: met.TotalIn(), AvrIn: met.InMetric.APS(),
			SumOut: met.TotalOut(), AvrOut: met.OutMetric.APS()}
		mets[i] = rMet
	}

	result.Peers = mets
}

// Blockchain handle rpc request blockchain. It has no additional input parameter
func (rpc *AergoRPCService) Blockchain(ctx context.Context, in *types.Empty) (*types.BlockchainStatus, error) {
	//last, _ := rpc.ChainService.GetBestBlock()
	/*
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
	*/
	last, err := rpc.actorHelper.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	return &types.BlockchainStatus{
		BestBlockHash: last.BlockHash(),
		BestHeight:    last.GetHeader().GetBlockNo(),
	}, nil
}

// ListBlockMetadata handle rpc request
func (rpc *AergoRPCService) ListBlockMetadata(ctx context.Context, in *types.ListParams) (*types.BlockMetadataList, error) {
	blocks, err := rpc.getBlocks(ctx, in)
	if err != nil {
		return nil, err
	}
	var metas []*types.BlockMetadata
	for _, block := range blocks {
		metas = append(metas, &types.BlockMetadata{
			Hash:    block.BlockHash(),
			Header:  block.GetHeader(),
			Txcount: int32(len(block.GetBody().GetTxs())),
		})
	}
	return &types.BlockMetadataList{Blocks: metas}, nil
}

// ListBlockHeaders (Deprecated) handle rpc request listblocks
func (rpc *AergoRPCService) ListBlockHeaders(ctx context.Context, in *types.ListParams) (*types.BlockHeaderList, error) {
	blocks, err := rpc.getBlocks(ctx, in)
	if err != nil {
		return nil, err
	}
	for _, block := range blocks {
		block.Body = nil
	}
	return &types.BlockHeaderList{Blocks: blocks}, nil
}

func (rpc *AergoRPCService) getBlocks(ctx context.Context, in *types.ListParams) ([]*types.Block, error) {
	var maxFetchSize uint32
	// TODO refactor with almost same code is in p2pcmdblock.go
	if in.Size > uint32(1000) {
		maxFetchSize = uint32(1000)
	} else {
		maxFetchSize = in.Size
	}
	idx := uint32(0)
	hashes := make([][]byte, 0, maxFetchSize)
	blocks := make([]*types.Block, 0, maxFetchSize)
	var err error
	if len(in.Hash) > 0 {
		hash := in.Hash
		for idx < maxFetchSize {
			foundBlock, ok := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
				&message.GetBlock{BlockHash: hash}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#1"))
			if !ok || nil == foundBlock {
				break
			}
			hashes = append(hashes, foundBlock.BlockHash())
			blocks = append(blocks, foundBlock)
			idx++
			hash = foundBlock.Header.PrevBlockHash
			if len(hash) == 0 {
				break
			}
		}
		if in.Asc || in.Offset != 0 {
			err = errors.New("Has unsupported param")
		}
	} else {
		end := types.BlockNo(0)
		start := types.BlockNo(in.Height) - types.BlockNo(in.Offset)
		if start >= types.BlockNo(maxFetchSize) {
			end = start - types.BlockNo(maxFetchSize-1)
		}
		if in.Asc {
			for i := end; i <= start; i++ {
				foundBlock, ok := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
					&message.GetBlockByNo{BlockNo: i}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#2"))
				if !ok || nil == foundBlock {
					break
				}
				hashes = append(hashes, foundBlock.BlockHash())
				blocks = append(blocks, foundBlock)
				idx++
			}
		} else {
			for i := start; i >= end; i-- {
				foundBlock, ok := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
					&message.GetBlockByNo{BlockNo: i}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#2"))
				if !ok || nil == foundBlock {
					break
				}
				hashes = append(hashes, foundBlock.BlockHash())
				blocks = append(blocks, foundBlock)
				idx++
			}
		}
	}
	return blocks, err
}

func (rpc *AergoRPCService) BroadcastToListBlockStream(block *types.Block) error {
	var err error
	rpc.blockStreamLock.RLock()
	for _, stream := range rpc.blockStream {
		if stream != nil {
			err = stream.Send(block)
		}
	}
	rpc.blockStreamLock.RUnlock()
	if err != nil {
		return err
	}
	return nil
}

func (rpc *AergoRPCService) BroadcastToListBlockMetadataStream(meta *types.BlockMetadata) error {
	var err error
	rpc.blockMetadataStreamLock.RLock()
	for _, stream := range rpc.blockMetadataStream {
		if stream != nil {
			err = stream.Send(meta)
		}
	}
	rpc.blockMetadataStreamLock.RUnlock()
	if err != nil {
		return err
	}
	return nil
}

// real-time streaming most recent block header
func (rpc *AergoRPCService) ListBlockStream(in *types.Empty, stream types.AergoRPCService_ListBlockStreamServer) error {
	rpc.blockStreamLock.Lock()
	streamId := len(rpc.blockStream)
	rpc.blockStream = append(rpc.blockStream, stream)
	rpc.blockStreamLock.Unlock()

	for {
		select {
		case <-stream.Context().Done():
			rpc.blockStreamLock.Lock()
			rpc.blockStream[streamId] = nil
			if len(rpc.blockStream) > 1024 {
				for i := 0; i < len(rpc.blockStream); i++ {
					if rpc.blockStream[i] == nil {
						rpc.blockStream = append(rpc.blockStream[:i], rpc.blockStream[i+1:]...)
						i--
						break
					}
				}
			}
			rpc.blockStreamLock.Unlock()
			return nil
		}
	}
}

func (rpc *AergoRPCService) ListBlockMetadataStream(in *types.Empty, stream types.AergoRPCService_ListBlockMetadataStreamServer) error {
	rpc.blockMetadataStreamLock.Lock()
	streamId := len(rpc.blockMetadataStream)
	rpc.blockMetadataStream = append(rpc.blockMetadataStream, stream)
	rpc.blockMetadataStreamLock.Unlock()

	for {
		select {
		case <-stream.Context().Done():
			rpc.blockMetadataStreamLock.Lock()
			rpc.blockMetadataStream[streamId] = nil
			if len(rpc.blockMetadataStream) > 1024 {
				for i := 0; i < len(rpc.blockMetadataStream); i++ {
					if rpc.blockMetadataStream[i] == nil {
						rpc.blockMetadataStream = append(rpc.blockMetadataStream[:i], rpc.blockMetadataStream[i+1:]...)
						i--
						break
					}
				}
			}
			rpc.blockMetadataStreamLock.Unlock()
			return nil
		}
	}
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
		return nil, status.Errorf(codes.InvalidArgument, "Received no bytes")
	}
	if len(in.Value) == 32 {
		result, err = rpc.hub.RequestFuture(message.ChainSvc, &message.GetBlock{BlockHash: in.Value},
			defaultActorTimeout, "rpc.(*AergoRPCService).GetBlock#2").Result()
	} else if len(in.Value) == 8 {
		number := uint64(binary.LittleEndian.Uint64(in.Value))
		result, err = rpc.hub.RequestFuture(message.ChainSvc, &message.GetBlockByNo{BlockNo: number},
			defaultActorTimeout, "rpc.(*AergoRPCService).GetBlock#1").Result()
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid input. Should be a 32 byte hash or up to 8 byte number.")
	}
	if err != nil {
		return nil, err
	}
	found, err := rpc.msgHelper.ExtractBlockFromResponse(result)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "Not found")
	}
	return found, nil
}

// GetTX handle rpc request gettx
func (rpc *AergoRPCService) GetTX(ctx context.Context, in *types.SingleBytes) (*types.Tx, error) {
	result, err := rpc.actorHelper.CallRequestDefaultTimeout(message.MemPoolSvc,
		&message.MemPoolExist{Hash: in.Value})
	if err != nil {
		return nil, err
	}
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

// SendTX try to fill the nonce, sign, hash in the transaction automatically and commit it
func (rpc *AergoRPCService) SendTX(ctx context.Context, tx *types.Tx) (*types.CommitResult, error) {
	getStateResult, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetState{Account: tx.Body.Account}, defaultActorTimeout, "rpc.(*AergoRPCService).SendTx").Result()
	if err != nil {
		return nil, err
	}
	getStateRsp, ok := getStateResult.(message.GetStateRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(getStateResult))
	}
	if getStateRsp.Err != nil {
		return nil, status.Errorf(codes.Internal, "internal error : %s", getStateRsp.Err.Error())
	}
	tx.Body.Nonce = getStateRsp.State.GetNonce() + 1

	signTxResult, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.SignTx{Tx: tx, Requester: getStateRsp.Account}, defaultActorTimeout, "rpc.(*AergoRPCService).SendTX")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	signTxRsp, ok := signTxResult.(*message.SignTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(signTxResult))
	}
	if signTxRsp.Err != nil {
		return nil, signTxRsp.Err
	}
	tx = signTxRsp.Tx
	memPoolPutResult, err := rpc.hub.RequestFuture(message.MemPoolSvc,
		&message.MemPoolPut{Tx: tx},
		defaultActorTimeout, "rpc.(*AergoRPCService).SendTX").Result()
	memPoolPutRsp, ok := memPoolPutResult.(*message.MemPoolPutRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(memPoolPutResult))
	}
	resultErr := memPoolPutRsp.Err
	if resultErr != nil {
		return &types.CommitResult{Hash: tx.Hash, Error: convertError(resultErr), Detail: resultErr.Error()}, err
	}
	return &types.CommitResult{Hash: tx.Hash, Error: convertError(resultErr)}, err
}

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
	futures := make([]*actor.Future, len(in.Txs))
	results := &types.CommitResultList{Results: rs}
	//results := &types.CommitResultList{}
	cnt := 0

	for i, tx := range in.Txs {
		hash := tx.Hash
		var r types.CommitResult
		r.Hash = hash

		calculated := tx.CalculateTxHash()

		if !bytes.Equal(hash, calculated) {
			r.Error = types.CommitStatus_TX_INVALID_HASH
		}
		results.Results[i] = &r
		cnt++

		//send tx message to mempool
		f := rpc.hub.RequestFuture(message.MemPoolSvc,
			&message.MemPoolPut{Tx: tx},
			defaultActorTimeout, "rpc.(*AergoRPCService).CommitTX")
		futures[i] = f
	}
	for i, future := range futures {
		result, err := future.Result()
		if err != nil {
			return nil, err
		}
		rsp, ok := result.(*message.MemPoolPutRsp)
		if !ok {
			err = status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
		} else {
			err = rsp.Err
		}
		results.Results[i].Error = convertError(err)
		if err != nil {
			results.Results[i].Detail = err.Error()
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
	return rsp.State, rsp.Err
}

// GetStateAndProof handle rpc request getstateproof
func (rpc *AergoRPCService) GetStateAndProof(ctx context.Context, in *types.AccountAndRoot) (*types.StateProof, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetStateAndProof{Account: in.Account, Root: in.Root, Compressed: in.Compressed}, defaultActorTimeout, "rpc.(*AergoRPCService).GetStateAndProof").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetStateAndProofRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.StateProof, rsp.Err
}

// CreateAccount handle rpc request newaccount
func (rpc *AergoRPCService) CreateAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.CreateAccount{Passphrase: in.Passphrase}, defaultActorTimeout, "rpc.(*AergoRPCService).CreateAccount")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	/*
		//same code but not good at folding in editor
		switch err {
		case nil:
		case component.ErrHubUnregistered:
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		default:
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	*/

	rsp, ok := result.(*message.CreateAccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, nil
	/*
		//it's better?
		switch rsp := result.(type) {
		case *message.CreateAccountRsp:
			return rsp.Accounts, nil
		default:
			return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
		}
	*/
}

// GetAccounts handle rpc request getaccounts
func (rpc *AergoRPCService) GetAccounts(ctx context.Context, in *types.Empty) (*types.AccountList, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.GetAccounts{}, defaultActorTimeout, "rpc.(*AergoRPCService).GetAccounts")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.GetAccountsRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Accounts, nil
}

// LockAccount handle rpc request lockaccount
func (rpc *AergoRPCService) LockAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.LockAccount{Account: in.Account, Passphrase: in.Passphrase},
		defaultActorTimeout, "rpc.(*AergoRPCService).LockAccount")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.AccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, rsp.Err
}

// UnlockAccount handle rpc request unlockaccount
func (rpc *AergoRPCService) UnlockAccount(ctx context.Context, in *types.Personal) (*types.Account, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.UnlockAccount{Account: in.Account, Passphrase: in.Passphrase},
		defaultActorTimeout, "rpc.(*AergoRPCService).UnlockAccount")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.AccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, rsp.Err
}

func (rpc *AergoRPCService) ImportAccount(ctx context.Context, in *types.ImportFormat) (*types.Account, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.ImportAccount{Wif: in.Wif.Value, OldPass: in.Oldpass, NewPass: in.Newpass},
		defaultActorTimeout, "rpc.(*AergoRPCService).ImportAccount")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.ImportAccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Account, rsp.Err
}

func (rpc *AergoRPCService) ExportAccount(ctx context.Context, in *types.Personal) (*types.SingleBytes, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.ExportAccount{Account: in.Account, Pass: in.Passphrase},
		defaultActorTimeout, "rpc.(*AergoRPCService).ExportAccount")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.ExportAccountRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return &types.SingleBytes{Value: rsp.Wif}, rsp.Err
}

// SignTX handle rpc request signtx
func (rpc *AergoRPCService) SignTX(ctx context.Context, in *types.Tx) (*types.Tx, error) {
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.SignTx{Tx: in}, defaultActorTimeout, "rpc.(*AergoRPCService).SignTX")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.SignTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Tx, rsp.Err
}

// VerifyTX handle rpc request verifytx
func (rpc *AergoRPCService) VerifyTX(ctx context.Context, in *types.Tx) (*types.VerifyResult, error) {
	//TODO : verify without account service
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.VerifyTx{Tx: in}, defaultActorTimeout, "rpc.(*AergoRPCService).VerifyTX")
	if err != nil {
		if err == component.ErrHubUnregistered {
			return nil, status.Errorf(codes.Unavailable, "Unavailable personal feature")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	rsp, ok := result.(*message.VerifyTxRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	ret := &types.VerifyResult{Tx: rsp.Tx}
	if rsp.Err == types.ErrSignNotMatch {
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
	rsp, ok := result.(*message.GetPeersRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}

	ret := &types.PeerList{Peers: []*types.Peer{}}
	for i, state := range rsp.States {
		peer := &types.Peer{Address: rsp.Peers[i], State: int32(state), Bestblock: rsp.LastBlks[i]}
		ret.Peers = append(ret.Peers, peer)
	}

	return ret, nil
}

// NodeState handle rpc request nodestate
func (rpc *AergoRPCService) NodeState(ctx context.Context, in *types.NodeReq) (*types.SingleBytes, error) {
	timeout := int64(binary.LittleEndian.Uint64(in.Timeout))
	component := string(in.Component)

	logger.Debug().Str("comp", component).Int64("timeout", timeout).Msg("nodestate")

	statics, err := rpc.hub.Statistics(time.Duration(timeout)*time.Second, component)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(statics, "", "\t")
	if err != nil {
		return nil, err
	}
	return &types.SingleBytes{Value: data}, nil
}

//GetVotes handle rpc request getvotes
func (rpc *AergoRPCService) GetVotes(ctx context.Context, in *types.SingleBytes) (*types.VoteList, error) {
	var number int
	var err error
	var result interface{}

	if len(in.Value) == 8 {
		number = int(binary.LittleEndian.Uint64(in.Value))
		result, err = rpc.hub.RequestFuture(message.ChainSvc,
			&message.GetElected{N: number}, defaultActorTimeout, "rpc.(*AergoRPCService).GetVote").Result()
	} else if len(in.Value) == types.AddressLength {
		result, err = rpc.hub.RequestFuture(message.ChainSvc,
			&message.GetVote{Addr: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetVote").Result()
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "Only support count parameter")
	}
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetVoteRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Top, rsp.Err
}

//GetStaking handle rpc request getstaking
func (rpc *AergoRPCService) GetStaking(ctx context.Context, in *types.SingleBytes) (*types.Staking, error) {
	var err error
	var result interface{}

	if len(in.Value) <= types.AddressLength {
		result, err = rpc.hub.RequestFuture(message.ChainSvc,
			&message.GetStaking{Addr: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetStaking").Result()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "Only support valid address")
	}
	rsp, ok := result.(*message.GetStakingRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Staking, rsp.Err
}

func (rpc *AergoRPCService) GetNameInfo(ctx context.Context, in *types.Name) (*types.NameInfo, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetNameInfo{Name: in.Name}, defaultActorTimeout, "rpc.(*AergoRPCService).GetName").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetNameInfoRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Owner, nil
}

func (rpc *AergoRPCService) GetReceipt(ctx context.Context, in *types.SingleBytes) (*types.Receipt, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetReceipt{TxHash: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetReceipt").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetReceiptRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Receipt, rsp.Err
}

func (rpc *AergoRPCService) GetABI(ctx context.Context, in *types.SingleBytes) (*types.ABI, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetABI{Contract: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetABI").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetABIRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.ABI, rsp.Err
}

func (rpc *AergoRPCService) QueryContract(ctx context.Context, in *types.Query) (*types.SingleBytes, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetQuery{Contract: in.ContractAddress, Queryinfo: in.Queryinfo}, defaultActorTimeout, "rpc.(*AergoRPCService).QueryContract").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetQueryRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return &types.SingleBytes{Value: rsp.Result}, rsp.Err
}

// QueryContractState queries the state of a contract state variable without executing a contract function.
func (rpc *AergoRPCService) QueryContractState(ctx context.Context, in *types.StateQuery) (*types.StateQueryProof, error) {
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetStateQuery{ContractAddress: in.ContractAddress, VarName: in.VarName, VarIndex: in.VarIndex, Root: in.Root, Compressed: in.Compressed}, defaultActorTimeout, "rpc.(*AergoRPCService).GetStateQuery").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(message.GetStateQueryRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Result, rsp.Err
}

func toTimestamp(time time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: time.Unix(),
		Nanos:   int32(time.Nanosecond())}
}

func fromTimestamp(timestamp *timestamp.Timestamp) time.Time {
	return time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
}
