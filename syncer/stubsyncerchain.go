package syncer

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/pkg/errors"
	"time"
)

//StubSyncer receive Syncer, P2P, Chain Service actor message
type StubBlockChain struct {
	best   int
	hashes []([]byte)
	blocks []*types.Block

	bestBlock *types.Block
}

type StubPeer struct {
	addr    *types.PeerAddress
	lastBlk *types.NewBlockNotice
	state   types.PeerState

	blockChain *StubBlockChain

	blockFetched bool //check if called while testing

	timeDelaySec time.Duration
}

var (
	TestMaxBlockFetchSize = 2
	TestMaxHashReqSize    = uint64(3)
)

var (
	ErrNotExistHash  = errors.New("not exist hash")
	ErrNotExistBlock = errors.New("not exist block of the hash")
)

func NewStubBlockChain() *StubBlockChain {
	tchain := &StubBlockChain{best: -1}

	tchain.hashes = make([][]byte, 10240)
	tchain.blocks = make([]*types.Block, 10240)

	return tchain
}

func (tchain *StubBlockChain) genAddBlock() {
	newBlock := types.NewBlock(tchain.bestBlock, nil, nil, nil, nil, time.Now().UnixNano())
	tchain.addBlock(newBlock)

	time.Sleep(time.Nanosecond * 3)
}

func (tchain *StubBlockChain) addBlock(newBlock *types.Block) error {
	if newBlock.BlockNo() != uint64(tchain.best+1) {
		return chain.ErrBlockOrphan
	}
	tchain.best += 1
	tchain.hashes[tchain.best] = newBlock.BlockHash()
	tchain.blocks[tchain.best] = newBlock
	tchain.bestBlock = newBlock

	return nil
}

func (tchain *StubBlockChain) GetHashes(prevInfo *types.BlockInfo, count uint64) ([]message.BlockHash, error) {
	if tchain.best < int(prevInfo.No+count) {
		return nil, ErrNotExistHash
	}

	start := prevInfo.No + 1
	resHashes := tchain.hashes[start : start+count]

	blkHashes := make([]message.BlockHash, 0)
	for _, hash := range resHashes {
		blkHashes = append(blkHashes, hash)
	}

	return blkHashes, nil
}

func (tchain *StubBlockChain) GetBlockInfo(no uint64) *types.BlockInfo {
	return &types.BlockInfo{tchain.hashes[no], no}
}

func (tchain *StubBlockChain) GetBlockByNo(no uint64) *types.Block {
	return tchain.blocks[no]
}

func (tchain *StubBlockChain) GetBlocks(hashes []message.BlockHash) ([]*types.Block, error) {
	startNo := -1

	for i, block := range tchain.blocks {
		if bytes.Equal(block.GetHash(), hashes[0]) {
			startNo = i
			break
		}
	}

	if startNo == -1 {
		return nil, ErrNotExistBlock
	}

	resultBlocks := make([]*types.Block, 0)
	i := startNo
	for _, hash := range hashes {
		if !bytes.Equal(tchain.blocks[i].GetHash(), hash) {
			return nil, ErrNotExistBlock
		}

		resultBlocks = append(resultBlocks, tchain.blocks[i])
		i++
	}

	return resultBlocks, nil
}

func (tchain *StubBlockChain) GetBestBlock() (*types.Block, error) {
	return tchain.bestBlock, nil
}

func (tchain *StubBlockChain) GetBlock(blockHash []byte) (*types.Block, error) {
	for _, block := range tchain.blocks {
		if bytes.Equal(block.GetHash(), blockHash) {
			return block, nil
			break
		}
	}

	return nil, ErrNotExistBlock
}

func (tchain *StubBlockChain) GetHashByNo(blockNo types.BlockNo) ([]byte, error) {
	if uint64(len(tchain.hashes)) <= blockNo {
		return nil, ErrNotExistHash
	}

	return tchain.hashes[blockNo], nil
}

//TODO refactoring with chain.getAnchorsNew()
func (tchain *StubBlockChain) GetAnchors() (chain.ChainAnchor, types.BlockNo, error) {
	//from top : 8 * 32 = 256
	anchors := make(chain.ChainAnchor, 0)
	cnt := chain.MaxAnchors
	logger.Debug().Msg("get anchors")

	bestBlock, _ := tchain.GetBestBlock()
	blkNo := bestBlock.BlockNo()
	var lastNo types.BlockNo
LOOP:
	for i := 0; i < cnt; i++ {
		blockHash, err := tchain.GetHashByNo(blkNo)
		if err != nil {
			logger.Info().Msg("assertion - hash get failed")
			// assertion!
			return nil, 0, err
		}

		anchors = append(anchors, blockHash)
		lastNo = blkNo

		logger.Debug().Uint64("no", blkNo).Msg("anchor added")

		switch {
		case blkNo == 0:
			break LOOP
		case blkNo < chain.Skip:
			blkNo = 0
		default:
			blkNo -= chain.Skip
		}
	}

	return anchors, lastNo, nil
}

func (tchain *StubBlockChain) GetAncestorWithHashes(hashes [][]byte) *types.BlockInfo {
	for _, hash := range hashes {
		for j, chainHash := range tchain.hashes {
			if bytes.Equal(hash, chainHash) {
				return &types.BlockInfo{Hash: chainHash, No: uint64(j)}
			}
		}
	}

	return nil
}

func (tchain *StubBlockChain) Rollback(ancestor *types.BlockInfo) {
	prevBest := tchain.best
	tchain.best = int(ancestor.No)
	tchain.bestBlock = tchain.blocks[tchain.best]

	logger.Debug().Int("prev", prevBest).Int("best", tchain.best).Msg("test local chain is rollbacked")
}

func initStubBlockChain(prefixChain []*types.Block, genCount int) *StubBlockChain {
	newChain := NewStubBlockChain()

	//load initial blocks
	for _, block := range prefixChain {
		newChain.addBlock(block)
	}

	for i := 0; i < genCount; i++ {
		newChain.genAddBlock()
	}

	return newChain
}

func NewStubPeer(idx int, lastNo uint64, blockChain *StubBlockChain) *StubPeer {
	stubPeer := &StubPeer{}

	peerIDBytes := []byte(fmt.Sprintf("peer-%d", idx))
	stubPeer.addr = &types.PeerAddress{PeerID: peerIDBytes}
	stubPeer.lastBlk = &types.NewBlockNotice{BlockNo: lastNo}
	stubPeer.state = types.RUNNING

	stubPeer.blockChain = blockChain

	return stubPeer
}
