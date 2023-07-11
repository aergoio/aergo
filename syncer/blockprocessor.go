package syncer

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
)

type BlockProcessor struct {
	compRequester component.IComponentRequester //for communicate with other service

	blockFetcher *BlockFetcher

	curConnRequest *ConnectTask

	connQueue []*ConnectTask

	prevBlock *types.Block
	curBlock  *types.Block

	targetBlockNo types.BlockNo
	name          string
}

type ConnectTask struct {
	FromPeer types.PeerID
	Blocks   []*types.Block
	firstNo  types.BlockNo
	cur      int
}

func NewBlockProcessor(compRequester component.IComponentRequester, blockFetcher *BlockFetcher, ancestor *types.Block,
	targetNo types.BlockNo) *BlockProcessor {
	return &BlockProcessor{
		compRequester: compRequester,
		blockFetcher:  blockFetcher,
		prevBlock:     ancestor,
		targetBlockNo: targetNo,
		name:          NameBlockProcessor,
	}
}

func (bproc *BlockProcessor) run(msg interface{}) error {
	//TODO in test mode, if syncer receives invalid messages, syncer stop with panic()
	switch msg.(type) {
	case *message.GetBlockChunksRsp:
		if err := bproc.GetBlockChunkRsp(msg.(*message.GetBlockChunksRsp)); err != nil {
			return err
		}
	case *message.AddBlockRsp:
		if err := bproc.AddBlockResponse(msg.(*message.AddBlockRsp)); err != nil {
			return err
		}

		chain.TestDebugger.Check(chain.DEBUG_SYNCER_CRASH, 2, nil)
	default:
		return fmt.Errorf("invalid msg type:%T", msg)
	}

	return nil
}

func (bproc *BlockProcessor) isValidResponse(msg interface{}) error {
	validateBlockChunksRsp := func(msg *message.GetBlockChunksRsp) error {
		var prev []byte
		blocks := msg.Blocks

		if msg.Err != nil {
			logger.Error().Err(msg.Err).Msg("GetBlockChunksRsp has error")
			return msg.Err
		}

		if blocks == nil || len(blocks) == 0 {
			logger.Error().Err(msg.Err).Str("peer", p2putil.ShortForm(msg.ToWhom)).Msg("GetBlockChunksRsp is empty")
			return &ErrSyncMsg{msg: msg, str: "blocks is empty"}
		}

		for _, block := range blocks {
			if prev != nil && !bytes.Equal(prev, block.GetHeader().GetPrevBlockHash()) {
				logger.Error().Str("peer", p2putil.ShortForm(msg.ToWhom)).Msg("GetBlockChunksRsp hashes inconsistent")
				return &ErrSyncMsg{msg: msg, str: "blocks hash not matched"}
			}

			prev = block.GetHash()
		}
		return nil
	}

	validateAddBlockRsp := func(msg *message.AddBlockRsp) error {
		if msg.Err != nil {
			return msg.Err
		}

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
	if err := bproc.isValidResponse(msg); err != nil {
		return bproc.GetBlockChunkRspError(msg, err)
	}

	bf := bproc.blockFetcher

	logger.Debug().Str("peer", p2putil.ShortForm(msg.ToWhom)).Uint64("startNo", msg.Blocks[0].GetHeader().BlockNo).Int("count", len(msg.Blocks)).Msg("received GetBlockChunkRsp")

	task, err := bf.findFinished(msg, false)
	if err != nil {
		//TODO invalid peer
		logger.Error().Str("peer", p2putil.ShortForm(msg.ToWhom)).
			Int("count", len(msg.Blocks)).
			Str("from", enc.ToString(msg.Blocks[0].GetHash())).
			Str("to", enc.ToString(msg.Blocks[len(msg.Blocks)-1].GetHash())).
			Msg("dropped unknown block response message")
		return nil
	}

	bf.pushFreePeer(task.syncPeer)

	bf.stat.setMaxChunkRsp(msg.Blocks[len(msg.Blocks)-1])

	bproc.addConnectTask(msg)

	return nil
}

func (bproc *BlockProcessor) GetBlockChunkRspError(msg *message.GetBlockChunksRsp, err error) error {
	bf := bproc.blockFetcher

	logger.Error().Err(err).Str("peer", p2putil.ShortForm(msg.ToWhom)).Msg("receive GetBlockChunksRsp with error message")

	task, err := bf.findFinished(msg, true)
	if err != nil {
		//TODO invalid peer
		logger.Error().Err(err).Str("peer", p2putil.ShortForm(msg.ToWhom)).Msg("dropped unknown block error message")
		return nil
	}

	if err := bf.processFailedTask(task, false); err != nil {
		return err
	}

	return nil
}

func (bproc *BlockProcessor) AddBlockResponse(msg *message.AddBlockRsp) error {
	if err := bproc.isValidResponse(msg); err != nil {
		logger.Info().Err(err).Uint64("no", msg.BlockNo).Str("hash", enc.ToString(msg.BlockHash)).Msg("block connect failed")
		return err
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
		logger.Info().Msg("succeed to add last block, request stopping syncer")
		stopSyncer(bproc.compRequester, bproc.blockFetcher.GetSeq(), bproc.name, nil)
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

	bproc.compRequester.RequestTo(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: nil, IsSync: true})
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
		logger.Debug().Msg("connect queue is empty. so wait new connect task")
		return nil
	}

	//check if first task is next block
	firstReq := sortedList[0]
	if bproc.prevBlock != nil &&
		firstReq.firstNo != (bproc.prevBlock.BlockNo()+1) {
		logger.Debug().Uint64("first", firstReq.firstNo).Uint64("prev", bproc.prevBlock.BlockNo()).Msg("next block is not fetched yet")
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
