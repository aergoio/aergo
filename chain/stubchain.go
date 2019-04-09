package chain

import (
	"bytes"
	"errors"
	"time"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
)

//StubSyncer receive Syncer, P2P, Chain Service actor message
type StubBlockChain struct {
	Best   int
	Hashes []([]byte)
	Blocks []*types.Block

	BestBlock *types.Block
}

var (
	ErrNotExistHash  = errors.New("not exist hash")
	ErrNotExistBlock = errors.New("not exist block of the hash")
)

func NewStubBlockChain() *StubBlockChain {
	tchain := &StubBlockChain{Best: -1}

	tchain.Hashes = make([][]byte, 10240)
	tchain.Blocks = make([]*types.Block, 10240)

	return tchain
}

func (tchain *StubBlockChain) GenAddBlock() {
	var prevBlockRootHash []byte
	if tchain.BestBlock != nil {
		prevBlockRootHash = tchain.BestBlock.GetHeader().BlocksRootHash
	}

	newBlock := types.NewBlock(tchain.BestBlock, prevBlockRootHash, nil, nil, nil, time.Now().UnixNano())
	tchain.AddBlock(newBlock)

	time.Sleep(time.Nanosecond * 3)
}

func (tchain *StubBlockChain) AddBlock(newBlock *types.Block) error {
	if newBlock.BlockNo() != uint64(tchain.Best+1) {
		return ErrBlockOrphan
	}
	tchain.Best += 1
	tchain.Hashes[tchain.Best] = newBlock.BlockHash()
	tchain.Blocks[tchain.Best] = newBlock
	tchain.BestBlock = newBlock

	return nil
}

func (tchain *StubBlockChain) GetHashes(prevInfo *types.BlockInfo, count uint64) ([]message.BlockHash, error) {
	if tchain.Best < int(prevInfo.No+count) {
		return nil, ErrNotExistHash
	}

	start := prevInfo.No + 1
	resHashes := tchain.Hashes[start : start+count]

	blkHashes := make([]message.BlockHash, 0)
	for _, hash := range resHashes {
		blkHashes = append(blkHashes, hash)
	}

	return blkHashes, nil
}

func (tchain *StubBlockChain) GetBlockInfo(no uint64) *types.BlockInfo {
	return &types.BlockInfo{tchain.Hashes[no], no}
}

func (tchain *StubBlockChain) GetBlockByNo(no uint64) *types.Block {
	return tchain.Blocks[no]
}

func (tchain *StubBlockChain) GetBlocks(hashes []message.BlockHash) ([]*types.Block, error) {
	startNo := -1

	for i, block := range tchain.Blocks {
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
		if !bytes.Equal(tchain.Blocks[i].GetHash(), hash) {
			return nil, ErrNotExistBlock
		}

		resultBlocks = append(resultBlocks, tchain.Blocks[i])
		i++
	}

	return resultBlocks, nil
}

func (tchain *StubBlockChain) GetGenesisInfo() *types.Genesis {
	// Not implemented. It should be implemented later if any test is related
	// to genesis info.
	return nil
}

func (tchain *StubBlockChain) GetConsensusInfo() string {
	return ""
}

func (tchain *StubBlockChain) GetChainStats() string {
	return ""
}

func (tchain *StubBlockChain) GetBestBlock() (*types.Block, error) {
	return tchain.BestBlock, nil
}

func (tchain *StubBlockChain) GetBlock(blockHash []byte) (*types.Block, error) {
	for _, block := range tchain.Blocks {
		if bytes.Equal(block.GetHash(), blockHash) {
			return block, nil
			break
		}
	}

	return nil, ErrNotExistBlock
}

func (tchain *StubBlockChain) GetHashByNo(blockNo types.BlockNo) ([]byte, error) {
	if uint64(len(tchain.Hashes)) <= blockNo {
		return nil, ErrNotExistHash
	}

	return tchain.Hashes[blockNo], nil
}

//TODO refactoring with getAnchorsNew()
func (tchain *StubBlockChain) GetAnchors() (ChainAnchor, types.BlockNo, error) {
	//from top : 8 * 32 = 256
	anchors := make(ChainAnchor, 0)
	cnt := MaxAnchors
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
		case blkNo < Skip:
			blkNo = 0
		default:
			blkNo -= Skip
		}
	}

	return anchors, lastNo, nil
}

func (tchain *StubBlockChain) GetAncestorWithHashes(hashes [][]byte) *types.BlockInfo {
	for _, hash := range hashes {
		for j, chainHash := range tchain.Hashes {
			if bytes.Equal(hash, chainHash) {
				return &types.BlockInfo{Hash: chainHash, No: uint64(j)}
			}
		}
	}

	return nil
}

func (tchain *StubBlockChain) Rollback(ancestor *types.BlockInfo) {
	prevBest := tchain.Best
	tchain.Best = int(ancestor.No)
	tchain.BestBlock = tchain.Blocks[tchain.Best]

	logger.Debug().Int("prev", prevBest).Int("Best", tchain.Best).Msg("test local chain is rollbacked")
}

func InitStubBlockChain(prefixChain []*types.Block, genCount int) *StubBlockChain {
	newChain := NewStubBlockChain()

	//load initial Blocks
	for _, block := range prefixChain {
		newChain.AddBlock(block)
	}

	for i := 0; i < genCount; i++ {
		newChain.GenAddBlock()
	}

	return newChain
}
