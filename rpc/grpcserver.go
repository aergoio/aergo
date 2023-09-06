/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package rpc

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/metric"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logger = log.NewLogger("rpc")
)

var (
	ErrUninitAccessor = errors.New("accessor is not initilized")

	//	ErrNotSupportedConsensus = errors.New("not supported by this consensus")
)

type EventStream struct {
	filter *types.FilterInfo
	stream types.AergoRPCService_ListEventStreamServer
}

// AergoRPCService implements GRPC server which is defined in rpc.proto
type AergoRPCService struct {
	hub               *component.ComponentHub
	actorHelper       p2pcommon.ActorService
	consensusAccessor consensus.ConsensusAccessor //TODO refactor with actorHelper
	msgHelper         message.Helper

	streamID                uint32
	blockStreamLock         sync.RWMutex
	blockStream             map[uint32]*ListBlockStream
	blockMetadataStreamLock sync.RWMutex
	blockMetadataStream     map[uint32]*ListBlockMetaStream

	eventStreamLock sync.RWMutex
	eventStream     map[*EventStream]*EventStream

	clientAuthLock sync.RWMutex
	clientAuthOn   bool
	clientAuth     map[string]Authentication
}

// FIXME remove redundant constants
const halfMinute = time.Second * 30
const defaultActorTimeout = time.Second * 3

var _ types.AergoRPCServiceServer = (*AergoRPCService)(nil)

func (rpc *AergoRPCService) SetConsensusAccessor(ca consensus.ConsensusAccessor) {
	if rpc == nil {
		return
	}

	rpc.consensusAccessor = ca
}

func (rpc *AergoRPCService) Metric(ctx context.Context, req *types.MetricsRequest) (*types.Metrics, error) {
	if err := rpc.checkAuth(ctx, ShowNode); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	ca := rpc.actorHelper.GetChainAccessor()
	last, err := ca.GetBestBlock()
	if err != nil {
		return nil, err
	}

	digest := sha256.New()
	digest.Write(last.GetHeader().GetChainID())
	bestChainIDHash := digest.Sum(nil)

	chainInfo, err := rpc.getChainInfo(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get chain info in blockchain")
		chainInfo = nil
	}
	return &types.BlockchainStatus{
		BestBlockHash:   last.BlockHash(),
		BestHeight:      last.GetHeader().GetBlockNo(),
		ConsensusInfo:   ca.GetConsensusInfo(),
		BestChainIdHash: bestChainIDHash,
		ChainInfo:       chainInfo,
	}, nil
}

// GetChainInfo handles a getchaininfo RPC request.
func (rpc *AergoRPCService) GetChainInfo(ctx context.Context, in *types.Empty) (*types.ChainInfo, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	return rpc.getChainInfo(ctx)
}

func (rpc *AergoRPCService) getChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	chainInfo := &types.ChainInfo{}

	if genesisInfo := rpc.actorHelper.GetChainAccessor().GetGenesisInfo(); genesisInfo != nil {
		ca := rpc.actorHelper.GetChainAccessor()
		last, err := ca.GetBestBlock()
		if err != nil {
			return nil, err
		}
		id := types.NewChainID()
		if err = id.Read(last.GetHeader().GetChainID()); err != nil {
			return nil, err
		}
		chainInfo.Id = &types.ChainId{
			Magic:     id.Magic,
			Public:    id.PublicNet,
			Mainnet:   id.MainNet,
			Consensus: id.Consensus,
			Version:   id.Version,
		}
		if totalBalance := genesisInfo.TotalBalance(); totalBalance != nil {
			chainInfo.Maxtokens = totalBalance.Bytes()
		}
	}

	cInfo, err := rpc.GetConsensusInfo(ctx, &types.Empty{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	chainInfo.BpNumber = uint32(len(cInfo.GetBps()))

	chainInfo.Maxblocksize = uint64(chain.MaxBlockSize())

	if consensus.IsDposName(chainInfo.Id.Consensus) {
		if minStaking, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.StakingMin); minStaking != nil {
			chainInfo.Stakingminimum = minStaking.Bytes()
		} else {
			return nil, err
		}
		if total, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.StakingTotal); total != nil {
			chainInfo.Totalstaking = total.Bytes()
		} else {
			return nil, err
		}
		if totalVotingPower, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.TotalVotingPower); totalVotingPower != nil {
			chainInfo.Totalvotingpower = totalVotingPower.Bytes()
		} else if err != nil {
			return nil, err
		}
		if votingReward, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.VotingReward); votingReward != nil {
			chainInfo.Votingreward = votingReward.Bytes()
		} else {
			return nil, err
		}
	}

	if namePrice, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.NamePrice); namePrice != nil {
		chainInfo.Nameprice = namePrice.Bytes()
	} else {
		return nil, err
	}

	if gasPrice, err := rpc.actorHelper.GetChainAccessor().GetSystemValue(types.GasPrice); gasPrice != nil {
		chainInfo.Gasprice = gasPrice.Bytes()
	} else {
		return nil, err
	}

	return chainInfo, nil
}

// ListBlockMetadata handle rpc request
func (rpc *AergoRPCService) ListBlockMetadata(ctx context.Context, in *types.ListParams) (*types.BlockMetadataList, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	blocks, err := rpc.getBlocks(ctx, in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var metas []*types.BlockMetadata
	for _, block := range blocks {
		metas = append(metas, block.GetMetadata())
	}
	return &types.BlockMetadataList{Blocks: metas}, nil
}

// ListBlockHeaders (Deprecated) handle rpc request listblocks
func (rpc *AergoRPCService) ListBlockHeaders(ctx context.Context, in *types.ListParams) (*types.BlockHeaderList, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
			foundBlock, futureErr := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
				&message.GetBlock{BlockHash: hash}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#1"))
			if nil != futureErr {
				if idx == 0 {
					err = futureErr
				}
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
				foundBlock, futureErr := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
					&message.GetBlockByNo{BlockNo: i}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#2"))
				if nil != futureErr {
					if i == end {
						err = futureErr
					}
					break
				}
				hashes = append(hashes, foundBlock.BlockHash())
				blocks = append(blocks, foundBlock)
				idx++
			}
		} else {
			for i := start; i >= end; i-- {
				foundBlock, futureErr := extractBlockFromFuture(rpc.hub.RequestFuture(message.ChainSvc,
					&message.GetBlockByNo{BlockNo: i}, defaultActorTimeout, "rpc.(*AergoRPCService).ListBlockHeaders#2"))
				if nil != futureErr {
					if i == start {
						err = futureErr
					}
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

func (rpc *AergoRPCService) BroadcastToListBlockStream(block *types.Block) {
	var err error
	rpc.blockStreamLock.RLock()
	defer rpc.blockStreamLock.RUnlock()
	for _, stream := range rpc.blockStream {
		if stream != nil {
			rpc.blockStreamLock.RUnlock()
			err = stream.Send(block)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to broadcast block stream")
			}
			rpc.blockStreamLock.RLock()
		}
	}
}

func (rpc *AergoRPCService) BroadcastToListBlockMetadataStream(meta *types.BlockMetadata) {
	var err error
	rpc.blockMetadataStreamLock.RLock()
	defer rpc.blockMetadataStreamLock.RUnlock()

	for _, stream := range rpc.blockMetadataStream {
		if stream != nil {
			rpc.blockMetadataStreamLock.RUnlock()
			err = stream.Send(meta)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to broadcast block meta stream")
			}
			rpc.blockMetadataStreamLock.RLock()
		}
	}
}

// ListBlockStream starts a stream of new blocks
func (rpc *AergoRPCService) ListBlockStream(in *types.Empty, stream types.AergoRPCService_ListBlockStreamServer) error {
	streamId := atomic.AddUint32(&rpc.streamID, 1)
	rpc.blockStreamLock.Lock()
	blockStream := NewListBlockStream(streamId, stream)
	rpc.blockStream[streamId] = blockStream
	// create goroutine for broadcast
	go blockStream.StartSend()
	rpc.blockStreamLock.Unlock()
	logger.Debug().Uint32("id", streamId).Msg("block stream added")

	// The stream will be terminated after returning this function
	for {
		select {
		case <-blockStream.awayChan: // server cut connection of bad client
			rpc.finishBlockStream(blockStream)
			logger.Debug().Uint32("id", streamId).Msg("block stream deleted by server")
			return nil
		case <-stream.Context().Done(): // client disconnected stream
			rpc.finishBlockStream(blockStream)
			logger.Debug().Uint32("id", streamId).Msg("block stream deleted")
			return nil
		}
	}
}

func (rpc *AergoRPCService) finishBlockStream(blockStream *ListBlockStream) {
	rpc.blockStreamLock.Lock()
	delete(rpc.blockStream, blockStream.id)
	blockStream.finishSend <- 0
	rpc.blockStreamLock.Unlock()
}

// ListBlockMetadataStream starts a stream of new blocks' metadata
func (rpc *AergoRPCService) ListBlockMetadataStream(in *types.Empty, stream types.AergoRPCService_ListBlockMetadataStreamServer) error {
	streamID := atomic.AddUint32(&rpc.streamID, 1)
	rpc.blockMetadataStreamLock.Lock()
	metadataStream := NewListBlockMetaStream(streamID, stream)
	rpc.blockMetadataStream[streamID] = metadataStream
	go metadataStream.StartSend()
	rpc.blockMetadataStreamLock.Unlock()
	logger.Debug().Uint32("id", streamID).Msg("block meta stream added")

	// The stream will be terminated after returning this function
	for {
		select {
		case <-metadataStream.awayChan: // server cut connection of bad client
			rpc.finishBlockMetadataStream(metadataStream)
			logger.Debug().Uint32("id", streamID).Msg("block meta stream deleted by server")
			return nil
		case <-stream.Context().Done(): // client disconnected stream
			rpc.finishBlockMetadataStream(metadataStream)
			logger.Debug().Uint32("id", streamID).Msg("block meta stream deleted")
			return nil
		}
	}
}

func (rpc *AergoRPCService) finishBlockMetadataStream(metadataStream *ListBlockMetaStream) {
	rpc.blockMetadataStreamLock.Lock()
	delete(rpc.blockMetadataStream, metadataStream.id)
	metadataStream.finishSend <- 0
	rpc.blockMetadataStreamLock.Unlock()
}

func extractBlockFromFuture(future *actor.Future) (*types.Block, error) {
	rawResponse, err := future.Result()
	if err != nil {
		return nil, err
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
		return nil, errors.New("Unsupported message type")
	}
	return extractBlock(blockRsp)
}

func extractBlock(from *message.GetBlockRsp) (*types.Block, error) {
	if nil != from.Err {
		return nil, from.Err
	}
	return from.Block, nil

}

// GetBlock handle rpc request getblock
func (rpc *AergoRPCService) GetBlock(ctx context.Context, in *types.SingleBytes) (*types.Block, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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

// GetBlockMetadata handle rpc request getblock
func (rpc *AergoRPCService) GetBlockMetadata(ctx context.Context, in *types.SingleBytes) (*types.BlockMetadata, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	block, err := rpc.GetBlock(ctx, in)
	if err != nil {
		return nil, err
	}
	meta := block.GetMetadata()
	return meta, nil
}

// GetBlockBody handle rpc request getblockbody
func (rpc *AergoRPCService) GetBlockBody(ctx context.Context, in *types.BlockBodyParams) (*types.BlockBodyPaged, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	block, err := rpc.GetBlock(ctx, &types.SingleBytes{Value: in.Hashornumber})
	if err != nil {
		return nil, err
	}
	body := block.GetBody()

	total := uint32(len(body.Txs))

	var fetchSize uint32
	if in.Paging.Size > uint32(1000) {
		fetchSize = uint32(1000)
	} else if in.Paging.Size == uint32(0) {
		fetchSize = 100
	} else {
		fetchSize = in.Paging.Size
	}

	offset := in.Paging.Offset
	if offset >= uint32(len(body.Txs)) {
		body.Txs = []*types.Tx{}
	} else {
		limit := offset + fetchSize
		if limit > uint32(len(body.Txs)) {
			limit = uint32(len(body.Txs))
		}
		body.Txs = body.Txs[offset:limit]
	}

	response := &types.BlockBodyPaged{
		Body:   body,
		Total:  total,
		Size:   fetchSize,
		Offset: offset,
	}
	return response, nil
}

// GetTX handle rpc request gettx
func (rpc *AergoRPCService) GetTX(ctx context.Context, in *types.SingleBytes) (*types.Tx, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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

// SendTX try to fill the nonce, sign, hash, chainIdHash in the transaction automatically and commit it
func (rpc *AergoRPCService) SendTX(ctx context.Context, tx *types.Tx) (*types.CommitResult, error) {
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
	if tx.Body.Nonce == 0 {
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
	}

	if tx.Body.ChainIdHash == nil {
		ca := rpc.actorHelper.GetChainAccessor()
		last, err := ca.GetBestBlock()
		if err != nil {
			return nil, err
		}
		tx.Body.ChainIdHash = common.Hasher(last.GetHeader().GetChainID())
	}

	signTxResult, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.SignTx{Tx: tx, Requester: tx.Body.Account}, defaultActorTimeout, "rpc.(*AergoRPCService).SendTX")
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
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
	if in.Txs == nil {
		return nil, status.Errorf(codes.InvalidArgument, "input tx is empty")
	}
	rpc.hub.Get(message.MemPoolSvc)
	p := newPutter(ctx, in.Txs, rpc.hub, defaultActorTimeout<<2)
	err := p.Commit()
	if err == nil {
		results := &types.CommitResultList{Results: p.rs}
		return results, nil
	} else {
		return nil, err
	}
}

// GetState handle rpc request getstate
func (rpc *AergoRPCService) GetState(ctx context.Context, in *types.SingleBytes) (*types.State, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
func (rpc *AergoRPCService) GetStateAndProof(ctx context.Context, in *types.AccountAndRoot) (*types.AccountProof, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ShowNode); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
	msg := &message.ImportAccount{OldPass: in.Oldpass, NewPass: in.Newpass}
	if in.Wif != nil {
		msg.Wif = in.Wif.Value
	} else if in.Keystore != nil {
		msg.Keystore = in.Keystore.Value
	} else {
		return nil, status.Errorf(codes.Internal, "require either wif or keystore contents")
	}
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		msg,
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

func (rpc *AergoRPCService) exportAccountWithFormat(ctx context.Context, in *types.Personal, asKeystore bool) (*types.SingleBytes, error) {
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFutureResult(message.AccountsSvc,
		&message.ExportAccount{Account: in.Account, Pass: in.Passphrase, AsKeystore: asKeystore},
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

func (rpc *AergoRPCService) ExportAccount(ctx context.Context, in *types.Personal) (*types.SingleBytes, error) {
	return rpc.exportAccountWithFormat(ctx, in, false)
}

func (rpc *AergoRPCService) ExportAccountKeystore(ctx context.Context, in *types.Personal) (*types.SingleBytes, error) {
	return rpc.exportAccountWithFormat(ctx, in, true)
}

// SignTX handle rpc request signtx
func (rpc *AergoRPCService) SignTX(ctx context.Context, in *types.Tx) (*types.Tx, error) {
	if err := rpc.checkAuth(ctx, WriteBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
func (rpc *AergoRPCService) GetPeers(ctx context.Context, in *types.PeersParams) (*types.PeerList, error) {
	if err := rpc.checkAuth(ctx, ShowNode); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.P2PSvc,
		&message.GetPeers{NoHidden: in.NoHidden, ShowSelf: in.ShowSelf}, halfMinute, "rpc.(*AergoRPCService).GetPeers").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetPeersRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}

	ret := &types.PeerList{Peers: make([]*types.Peer, 0, len(rsp.Peers))}
	for _, pi := range rsp.Peers {
		blkNotice := &types.NewBlockNotice{BlockHash: pi.LastBlockHash, BlockNo: pi.LastBlockNumber}
		peer := &types.Peer{Address: pi.Addr, State: int32(pi.State), Bestblock: blkNotice, LashCheck: pi.CheckTime.UnixNano(), Hidden: pi.Hidden, Selfpeer: pi.Self, Version: pi.Version, Certificates: pi.Certificates, AcceptedRole: pi.AcceptedRole}
		ret.Peers = append(ret.Peers, peer)
	}

	return ret, nil
}

// NodeState handle rpc request nodestate
func (rpc *AergoRPCService) NodeState(ctx context.Context, in *types.NodeReq) (*types.SingleBytes, error) {
	if err := rpc.checkAuth(ctx, ShowNode); err != nil {
		return nil, err
	}
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

// GetVotes handle rpc request getvotes
func (rpc *AergoRPCService) GetVotes(ctx context.Context, in *types.VoteParams) (*types.VoteList, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}

	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetElected{Id: in.GetId(), N: in.GetCount()}, defaultActorTimeout, "rpc.(*AergoRPCService).GetVote").Result()

	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetVoteRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Top, rsp.Err
}

func (rpc *AergoRPCService) GetAccountVotes(ctx context.Context, in *types.AccountAddress) (*types.AccountVoteInfo, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetVote{Addr: in.Value}, defaultActorTimeout, "rpc.(*AergoRPCService).GetAccountVote").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetAccountVoteRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Info, rsp.Err
}

// GetStaking handle rpc request getstaking
func (rpc *AergoRPCService) GetStaking(ctx context.Context, in *types.AccountAddress) (*types.Staking, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetNameInfo{Name: in.Name, BlockNo: in.BlockNo}, defaultActorTimeout, "rpc.(*AergoRPCService).GetName").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetNameInfoRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	if rsp.Err == types.ErrNameNotFound {
		return rsp.Owner, status.Errorf(codes.NotFound, rsp.Err.Error())
	}
	return rsp.Owner, rsp.Err
}

func (rpc *AergoRPCService) GetReceipt(ctx context.Context, in *types.SingleBytes) (*types.Receipt, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
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
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetStateQuery{ContractAddress: in.ContractAddress, StorageKeys: in.StorageKeys, Root: in.Root, Compressed: in.Compressed}, defaultActorTimeout, "rpc.(*AergoRPCService).GetStateQuery").Result()
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

func (rpc *AergoRPCService) ListEventStream(in *types.FilterInfo, stream types.AergoRPCService_ListEventStreamServer) error {
	err := in.ValidateCheck(0)
	if err != nil {
		return err
	}
	_, err = in.GetExArgFilter()
	if err != nil {
		return err
	}

	eventStream := &EventStream{in, stream}
	rpc.eventStreamLock.Lock()
	rpc.eventStream[eventStream] = eventStream
	rpc.eventStreamLock.Unlock()

	for {
		select {
		case <-eventStream.stream.Context().Done():
			rpc.eventStreamLock.Lock()
			delete(rpc.eventStream, eventStream)
			rpc.eventStreamLock.Unlock()
			return nil
		}
	}
}

func (rpc *AergoRPCService) BroadcastToEventStream(events []*types.Event) error {
	var err error
	rpc.eventStreamLock.RLock()
	defer rpc.eventStreamLock.RUnlock()

	for _, es := range rpc.eventStream {
		if es != nil {
			rpc.eventStreamLock.RUnlock()
			argFilter, _ := es.filter.GetExArgFilter()
			for _, event := range events {
				if event.Filter(es.filter, argFilter) {
					err = es.stream.Send(event)
					if err != nil {
						logger.Warn().Err(err).Msg("failed to broadcast block stream")
						break
					}
				}
			}
			rpc.eventStreamLock.RLock()
		}
	}
	return nil
}

func (rpc *AergoRPCService) ListEvents(ctx context.Context, in *types.FilterInfo) (*types.EventList, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.ListEvents{Filter: in}, defaultActorTimeout, "rpc.(*AergoRPCService).ListEvents").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.ListEventsRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return &types.EventList{Events: rsp.Events}, rsp.Err
}

func (rpc *AergoRPCService) GetServerInfo(ctx context.Context, in *types.KeyParams) (*types.ServerInfo, error) {
	if err := rpc.checkAuth(ctx, ShowNode); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.RPCSvc,
		&message.GetServerInfo{Categories: in.Key}, defaultActorTimeout, "rpc.(*AergoRPCService).GetServerInfo").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*types.ServerInfo)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp, nil
}

// GetConsensusInfo handle rpc request blockchain. It has no additional input parameter
func (rpc *AergoRPCService) GetConsensusInfo(ctx context.Context, in *types.Empty) (*types.ConsensusInfo, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	if rpc.consensusAccessor == nil {
		return nil, ErrUninitAccessor
	}

	return rpc.consensusAccessor.ConsensusInfo(), nil
}

// ChainStat handles rpc request chainstat.
func (rpc *AergoRPCService) ChainStat(ctx context.Context, in *types.Empty) (*types.ChainStats, error) {
	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	ca := rpc.actorHelper.GetChainAccessor()
	if ca == nil {
		return nil, ErrUninitAccessor
	}
	return &types.ChainStats{Report: ca.GetChainStats()}, nil
}

// GetEnterpriseConfig return aergo.enterprise configure values. key "ADMINS" is for getting register admin addresses and "ALL" is for getting all key list.
func (rpc *AergoRPCService) GetEnterpriseConfig(ctx context.Context, in *types.EnterpriseConfigKey) (*types.EnterpriseConfig, error) {
	genesis := rpc.actorHelper.GetChainAccessor().GetGenesisInfo()
	if genesis.PublicNet() {
		return nil, status.Error(codes.Unavailable, "not supported in public")
	}

	if err := rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}
	result, err := rpc.hub.RequestFuture(message.ChainSvc,
		&message.GetEnterpriseConf{Key: in.Key}, defaultActorTimeout, "rpc.(*AergoRPCService).GetEnterpiseConfig").Result()
	if err != nil {
		return nil, err
	}
	rsp, ok := result.(*message.GetEnterpriseConfRsp)
	if !ok {
		return nil, status.Errorf(codes.Internal, "internal type (%v) error", reflect.TypeOf(result))
	}
	return rsp.Conf, nil
}

func (rpc *AergoRPCService) GetConfChangeProgress(ctx context.Context, in *types.SingleBytes) (*types.ConfChangeProgress, error) {
	var (
		progress *types.ConfChangeProgress
		err      error
	)

	genesis := rpc.actorHelper.GetChainAccessor().GetGenesisInfo()
	if genesis.PublicNet() {
		return nil, status.Error(codes.Unavailable, "not supported in public")
	}

	if strings.ToLower(genesis.ConsensusType()) != consensus.ConsensusName[consensus.ConsensusRAFT] {
		return nil, status.Error(codes.Unavailable, "not supported if not raft consensus")
	}

	if err = rpc.checkAuth(ctx, ReadBlockChain); err != nil {
		return nil, err
	}

	if rpc.consensusAccessor == nil {
		return nil, ErrUninitAccessor
	}

	if len(in.Value) != 8 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid input. Request ID should be a 8 byte number.")
	}

	reqID := uint64(binary.LittleEndian.Uint64(in.Value))

	if progress, err = rpc.consensusAccessor.ConfChangeInfo(reqID); err != nil {
		return nil, err
	}

	if progress == nil {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	return progress, nil
}

func (rpc *AergoRPCService) Statistics() *map[string]interface{} {
	return &map[string]interface{}{
		"block":     len(rpc.blockStream),
		"blockMeta": len(rpc.blockMetadataStream),
		"event":     len(rpc.eventStream),
	}
}
