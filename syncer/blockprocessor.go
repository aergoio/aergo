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

	curConnRequest *ConnectTask

	connQueue []*ConnectTask

	prevBlock *types.Block
	curBlock  *types.Block

	targetBlockNo types.BlockNo
	name          string
}

type ConnectTask struct {
	FromPeer peer.ID
	Blocks   []*types.Block
	firstNo  types.BlockNo
	cur      int
}

func NewBlockProcessor(hub component.ICompRequester, blockFetcher *BlockFetcher, ancestor *types.Block,
	targetNo types.BlockNo) *BlockProcessor {
	return &BlockProcessor{
		hub:           hub,
		blockFetcher:  blockFetcher,
		prevBlock:     ancestor,
		targetBlockNo: targetNo,
		name:          NameBlockProcessor,
	}
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

	logger.Debug().Str("peer", msg.ToWhom.Pretty()).Uint64("startNo", msg.Blocks[0].GetHeader().BlockNo).Int("count", len(msg.Blocks)).Msg("received GetBlockCHunkRsp")

	task, err := bf.findFinished(msg)
	if err != nil {
		//TODO invalid peer
		logger.Error().Str("peer", msg.ToWhom.Pretty()).
			Int("count", len(msg.Blocks)).
			Str("from", enc.ToString(msg.Blocks[0].GetHash())).
			Str("to", enc.ToString(msg.Blocks[len(msg.Blocks)-1].GetHash())).
			Msg("dropped unknown block message")
		return nil
	}

	bf.pushFreePeer(task.syncPeer)

	bf.stat.setMaxChunkRsp(msg.Blocks[len(msg.Blocks)-1])

	bproc.addConnectTask(msg)

	return nil
}

func (bproc *BlockProcessor) GetBlockChunkRspError(msg *message.GetBlockChunksRsp) error {
	bf := bproc.blockFetcher

	logger.Error().Str("peer", msg.ToWhom.Pretty()).Msg("receive GetBlockChunksRsp with error message")

	task, err := bf.findFinished(msg)
	if err != nil {
		//TODO invalid peer
		logger.Error().Str("peer", msg.ToWhom.Pretty()).Msg("dropped unknown block error message")
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
			Msg("invalid add block response")
		return &ErrSyncMsg{msg: msg, str: "drop unknown add response"}
	}

	logger.Info().Uint64("no", msg.BlockNo).Str("hash", enc.ToString(msg.BlockHash)).Msg("block connect succeed")

	bproc.blockFetcher.stat.setLastAddBlock(curBlock)

	if curBlock.BlockNo() == bproc.targetBlockNo {
		logger.Info().Msg("connected last block, stop syncer")
		stopSyncer(bproc.hub, bproc.name, nil)
	}

	bproc.prevBlock = curBlock
	bproc.curBlock = nil

	block := bproc.getNextBlockToConnect()

	if block != nil {
		bproc.connectBlock(block)
	}

	return nil
}

func (bproc *BlockProcessor) addConnectTask(msg *message.GetBlockChunksRsp) {
	req := &ConnectTask{FromPeer: msg.ToWhom, Blocks: msg.Blocks, firstNo: msg.Blocks[0].GetHeader().BlockNo, cur: 0}

	logger.Debug().Uint64("firstno", req.firstNo).Int("count", len(req.Blocks)).Msg("add connect task to queue")

	bproc.pushToConnQueue(req)

	block := bproc.getNextBlockToConnect()

	if block != nil {
		bproc.connectBlock(block)
	}
}

func (bproc *BlockProcessor) getNextBlockToConnect() *types.Block {
	//already prev request is running, don't request any more
	if bproc.curBlock != nil {
		return nil
	}

	//request next block of current Request
	if bproc.curConnRequest != nil {
		req := bproc.curConnRequest
		req.cur++

		if req.cur >= len(req.Blocks) {
			logger.Debug().Msg("current connect task is finished")
			bproc.curConnRequest = nil
		}
	}

	//pop from pending request
	if bproc.curConnRequest == nil {
		nextReq := bproc.popFromConnQueue()
		if nextReq == nil {
			return nil
		}

		bproc.curConnRequest = nextReq
	}

	next := bproc.curConnRequest.cur
	nextBlock := bproc.curConnRequest.Blocks[next]

	logger.Debug().Uint64("no", nextBlock.GetHeader().BlockNo).Str("hash", nextBlock.ID()).
		Int("idx in req", next).Msg("next block to connect")

	bproc.curBlock = nextBlock

	return nextBlock
}

func (bproc *BlockProcessor) connectBlock(block *types.Block) {
	if block == nil {
		return
	}

	logger.Info().Uint64("no", block.GetHeader().BlockNo).
		Str("hash", enc.ToString(block.GetHash())).
		Msg("request connecting block to chainsvc")

	bproc.hub.Tell(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: nil, IsSync: true})
}

func (bproc *BlockProcessor) pushToConnQueue(newReq *ConnectTask) {
	sortedList := bproc.connQueue

	index := sort.Search(len(sortedList), func(i int) bool { return sortedList[i].firstNo > newReq.firstNo })
	sortedList = append(sortedList, &ConnectTask{})
	copy(sortedList[index+1:], sortedList[index:])
	sortedList[index] = newReq

	bproc.connQueue = sortedList

	logger.Info().Int("len", len(bproc.connQueue)).Uint64("firstno", newReq.firstNo).
		Str("firstHash", enc.ToString(newReq.Blocks[0].GetHash())).
		Msg("add new task to connect queue")
}

func (bproc *BlockProcessor) popFromConnQueue() *ConnectTask {
	sortedList := bproc.connQueue
	if len(sortedList) == 0 {
		logger.Info().Msg("connect queue is empty. so wait new connect task")
		return nil
	}

	newReq := sortedList[0]
	sortedList = sortedList[1:]
	bproc.connQueue = sortedList

	logger.Info().Int("len", len(sortedList)).Uint64("firstno", newReq.firstNo).
		Str("firstHash", enc.ToString(newReq.Blocks[0].GetHash())).
		Msg("pop task from connect queue")

	return newReq
}
