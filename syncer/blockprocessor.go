package syncer

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"sort"
)

type BlockProcessor struct {
	hub component.ICompRequester //for communicate with other service

	blockFetcher *BlockFetcher

	curConnRequest *ConnectRequest

	pendingConnect []*ConnectRequest

	prevBlock *types.Block
	curBlock  *types.Block

	targetBlockNo types.BlockNo
	name          string
}

type ConnectRequest struct {
	FromPeer peer.ID
	Blocks   []*types.Block
	firstNo  types.BlockNo
	cur      int
}

func (bproc *BlockProcessor) run(msg interface{}) error {
	//TODO in test mode, if syncer receives invalid messages, syncer stop with panic()

	if err := bproc.isValidResponse(msg); err != nil {
		logger.Error().Err(err).Msg("dropped invalid block message")
		return nil
	}

	switch msg.(type) {
	case *message.GetBlockChunksRsp:
		if err := bproc.GetBlockChunkRsp(msg.(*message.GetBlockChunksRsp)); err != nil {
			return err
		}
	case *message.AddBlockRsp:
		if err := bproc.AddBlockResponse(msg.(*message.AddBlockRsp)); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid msg type:%T", msg)
	}

	return nil
}

func (bproc *BlockProcessor) isValidResponse(msg interface{}) error {
	validateBlockChunksRsp := func(msg *message.GetBlockChunksRsp) error {
		var prev []byte
		blocks := msg.Blocks

		if msg.Err != nil && (blocks == nil || len(blocks) == 0) {
			return &ErrSyncMsg{msg: msg, str: "blocks is empty"}
		}

		for _, block := range blocks {
			if prev != nil && !bytes.Equal(prev, block.GetHeader().GetPrevBlockHash()) {
				return &ErrSyncMsg{msg: msg, str: "blocks hash not matched"}
			}

			prev = block.GetHash()
		}
		return nil
	}

	validateAddBlockRsp := func(msg *message.AddBlockRsp) error {
		if msg.BlockHash == nil {
			return &ErrSyncMsg{msg: msg, str: "invalid add block resonse"}
		}

		return nil
	}

	switch msg.(type) {
	case *message.GetBlockChunksRsp:
		if err := validateBlockChunksRsp(msg.(*message.GetBlockChunksRsp)); err != nil {
			return err
		}

	case *message.AddBlockRsp:
		if err := validateAddBlockRsp(msg.(*message.AddBlockRsp)); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid msg type:%T", msg)
	}

	return nil
}

func (bproc *BlockProcessor) GetBlockChunkRsp(msg *message.GetBlockChunksRsp) error {
	if msg.Err != nil {
		return bproc.GetBlockChunkRspError(msg)
	}

	bf := bproc.blockFetcher

	logger.Debug().Str("peer", string(msg.ToWhom)).Err(msg.Err).Msg("received GetBlockCHunkRsp")

	task, err := bf.findFinished(msg)
	if err != nil {
		//TODO invalid peer
		logger.Error().Str("peer", msg.ToWhom.String()).
			Int("count", len(msg.Blocks)).
			Str("from", enc.ToString(msg.Blocks[0].GetHash())).
			Str("to", enc.ToString(msg.Blocks[len(msg.Blocks)-1].GetHash())).
			Msg("dropped unknown block message")
		return nil
	}

	bf.pushFreePeer(task.syncPeer)

	bproc.addNewRequest(msg)

	return nil
}

func (bproc *BlockProcessor) GetBlockChunkRspError(msg *message.GetBlockChunksRsp) error {
	bf := bproc.blockFetcher

	logger.Error().Str("peer", msg.ToWhom.String()).Msg("receive GetBlockChunksRsp with error message")

	task, err := bf.findFinished(msg)
	if err != nil {
		//TODO invalid peer
		logger.Error().Str("peer", msg.ToWhom.String()).Msg("dropped unknown block error message")
		return nil
	}

	bf.processFailedTask(task, true)
	return nil
}

func (bproc *BlockProcessor) AddBlockResponse(msg *message.AddBlockRsp) error {
	if msg.Err != nil {
		logger.Error().Err(msg.Err).Msg("connect block failed")
		return msg.Err
	}

	curBlock := bproc.curBlock
	curNo := curBlock.GetHeader().BlockNo
	curHash := curBlock.GetHash()

	if curNo != msg.BlockNo || !bytes.Equal(curHash, msg.BlockHash) {
		logger.Error().Uint64("curNo", curNo).Uint64("msgNo", msg.BlockNo).
			Str("curHash", enc.ToString(curHash)).Str("msgHash", enc.ToString(msg.BlockHash)).
			Msg("error! unmatched add response")
		return &ErrSyncMsg{msg: msg, str: "drop unknown add response"}
	}

	if curBlock.BlockNo() == bproc.targetBlockNo {
		logger.Info().Msg("connected last block, stop syncer")
		stopSyncer(bproc.hub, bproc.name, nil)
	}

	bproc.prevBlock = curBlock

	block := bproc.getNextBlock()
	bproc.curBlock = block

	if block != nil {
		bproc.connectBlock(block)
	}

	return nil
}

func (bproc *BlockProcessor) addNewRequest(msg *message.GetBlockChunksRsp) {
	logger.Debug().Msg("add connect request to pending queue")

	req := &ConnectRequest{FromPeer: msg.ToWhom, Blocks: msg.Blocks, firstNo: msg.Blocks[0].GetHeader().BlockNo, cur: 0}

	bproc.pushToPending(req)

	block := bproc.getNextBlock()

	if block != nil {
		bproc.connectBlock(block)
	}
}

func (bproc *BlockProcessor) getNextBlock() *types.Block {
	//request next block of current Request
	if bproc.curConnRequest != nil {
		req := bproc.curConnRequest
		req.cur++

		if req.cur >= len(req.Blocks) {
			logger.Debug().Msg("current connrequest is finished")
			bproc.curConnRequest = nil
		}
	}

	//pop from pending request
	if bproc.curConnRequest == nil {
		nextReq := bproc.popFromPending()
		if nextReq == nil {
			return nil
		}

		bproc.curConnRequest = nextReq
	}

	next := bproc.curConnRequest.cur
	nextBlock := bproc.curConnRequest.Blocks[next]

	logger.Debug().Uint64("no", nextBlock.GetHeader().BlockNo).Str("hash", nextBlock.ID()).
		Int("idx in req", next).Msg("next block to connect")

	return nextBlock
}

func (bproc *BlockProcessor) connectBlock(block *types.Block) {
	if block == nil {
		return
	}

	logger.Info().Uint64("no", block.GetHeader().BlockNo).
		Str("hash", enc.ToString(block.GetHash())).
		Msg("send connect request to chainsvc")

	bproc.curBlock = block
	bproc.hub.Tell(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: nil})
}

func (bproc *BlockProcessor) pushToPending(newReq *ConnectRequest) {
	sortedList := bproc.pendingConnect

	index := sort.Search(len(sortedList), func(i int) bool { return sortedList[i].firstNo > newReq.firstNo })
	sortedList = append(sortedList, &ConnectRequest{})
	copy(sortedList[index+1:], sortedList[index:])
	sortedList[index] = newReq

	logger.Info().Int("len", len(bproc.pendingConnect)).Uint64("firstno", newReq.firstNo).
		Str("firstHash", enc.ToString(newReq.Blocks[0].GetHash())).
		Msg("add new request to pending queue")
}

func (bproc *BlockProcessor) popFromPending() *ConnectRequest {
	sortedList := bproc.pendingConnect
	if len(sortedList) == 0 {
		logger.Info().Msg("pending queue is empty. so wait new connect request")
		return nil
	}

	newReq := sortedList[0]
	sortedList = sortedList[1:]
	bproc.pendingConnect = sortedList

	logger.Info().Int("len", len(sortedList)).Uint64("firstno", newReq.firstNo).
		Str("firstHash", enc.ToString(newReq.Blocks[0].GetHash())).
		Msg("pop request from pending queue")

	return newReq
}
