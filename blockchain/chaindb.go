/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/pkg/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/gogo/protobuf/proto"
)

const (
	chainDBName = "chain"
)

var (
	latestKey = []byte(chainDBName + ".latest")
)

type ChainDB struct {
	consensus.ChainInfo

	latest types.BlockNo
	//	blocks []*types.Block
	store db.DB
}

func NewChainDB() *ChainDB {
	// logger.SetLevel("debug")
	cdb := &ChainDB{
		//blocks: []*types.Block{},
		latest: types.BlockNo(0),
	}

	return cdb
}

func (cdb *ChainDB) Init(dataDir string) error {
	if cdb.store == nil {
		cdb.store = state.InitDB(dataDir, chainDBName)
	}

	// load data
	if err := cdb.loadChainData(); err != nil {
		return err
	}
	// // if empty then create new genesis block
	// // if cdb.latest == 0 && len(cdb.blocks) == 0 {
	// blockKey := ItobU64(0)
	// blockHash := cdb.store.Get(blockKey)
	// if cdb.latest == 0 && (blockHash == nil || len(blockHash) == 0) {
	// 	cdb.generateGenesisBlock(seed)
	// }
	return nil
}

func (cdb *ChainDB) Close() {
	if cdb.store != nil {
		cdb.store.Close()
	}
	return
}

func (cdb *ChainDB) loadChainData() error {
	latestBytes := cdb.store.Get(latestKey)
	if latestBytes == nil || len(latestBytes) == 0 {
		return nil
	}
	latestInt := BtoiU64(latestBytes)
	/* TODO: just checking DB
	cdb.blocks = make([]*types.Block, latestInt+1)
	for i := uint32(0); i <= latestInt; i++ {
		blockKey := ItobU32(i)
		buf := types.Block{}
		err := cdb.loadData(blockKey, &buf)
		if err != nil {
			return err
		}
		bHash := buf.CalculateBlockHash()
		if buf.Hash == nil {
			buf.Hash = bHash
		} else if !bytes.Equal(buf.Hash, bHash) {
			return fmt.Errorf("invalid Block Hash: hash=%s, check=%s",
				EncodeB64(buf.Hash), EncodeB64(bHash))
		}
		for _, v := range buf.Body.Txs {
			tHash := v.CalculateTxHash()
			if v.Hash == nil {
				v.Hash = tHash
			} else if !bytes.Equal(v.Hash, tHash) {
				return fmt.Errorf("invalid Transaction Hash: hash=%s, check=%s",
					EncodeB64(v.Hash), EncodeB64(tHash))
			}
		}
		cdb.blocks[i] = &buf
	}
	*/
	cdb.setLatest(latestInt)

	// skips := true
	// for i, _ := range cdb.blocks {
	// 	if i > 3 && i+3 <= cdb.latest {
	// 		if skips {
	// 			skips = false
	// 			//logger.Info("  ...")
	// 		}
	// 		continue
	// 	}
	// 	//logger.Info("- loaded:", i, ToJSON(v))
	// }
	return nil
}
func (cdb *ChainDB) loadData(key []byte, pb proto.Message) error {
	buf := cdb.store.Get(key)
	if buf == nil || len(buf) == 0 {
		return fmt.Errorf("failed to load data: key=%v", key)
	}
	//logger.Debugf("  loadData: key=%d, len=%d, val=%s\n", Btoi(key), len(buf), EncodeB64(buf))
	err := proto.Unmarshal(buf, pb)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: key=%v, len=%d", key, len(buf))
	}
	//logger.Debug("  loaded: ", ToJSON(pb))
	return nil
}
func (cdb *ChainDB) generateGenesisBlock(seed int64) *types.Block {
	genesisBlock := types.NewBlock(nil, nil, 0)
	genesisBlock.Header.Timestamp = seed
	genesisBlock.Hash = genesisBlock.CalculateBlockHash()
	tx := cdb.store.NewTx(true)
	cdb.addBlock(&tx, genesisBlock)
	tx.Commit()
	logger.Info().Msg("generate Genesis Block")
	return genesisBlock
}

func (cdb *ChainDB) setLatest(newLatest types.BlockNo) {
	cdb.latest = newLatest
}

func (cdb *ChainDB) isNewBestBlock(block *types.Block) bool {
	blockNo := block.GetHeader().GetBlockNo()
	if blockNo > 0 && blockNo != cdb.latest+1 {
		return false
	}

	prevHash := block.GetHeader().GetPrevBlockHash()
	latestHash, err := cdb.getHashByNo(cdb.getBestBlockNo())
	if err != nil { //need assertion
		return false
	}

	isNewBest := bytes.Equal(prevHash, latestHash)
	if isNewBest {
		logger.Debug().Uint64("blkno", blockNo).Str("hash", block.ID()).Msg("new best block")
	}

	return isNewBest
}

type txInfo struct {
	blockHash []byte
	idx       int
}

// stor tx info to DB
func (cdb *ChainDB) addTx(dbtx *db.Transaction, tx *types.Tx, blockHash []byte, idx int) error {
	txidx := types.TxIdx{
		BlockHash: blockHash,
		Idx:       int32(idx),
	}
	txidxbytes, err := proto.Marshal(&txidx)
	if err != nil {
		return err
	}
	(*dbtx).Set(tx.Hash, txidxbytes)
	return nil
}

func (cdb *ChainDB) deleteTx(dbtx *db.Transaction, tx *types.Tx) {
	(*dbtx).Delete(tx.Hash)
}

// store block info to DB
func (cdb *ChainDB) addBlock(dbtx *db.Transaction, block *types.Block) error {
	blockNo := block.GetHeader().GetBlockNo()
	blockKey := ItobU64(blockNo)
	longest := true

	// FIXME: blockNo 0 exception handling
	// assumption: not an orphan
	// fork can be here
	if blockNo > 0 && blockNo != cdb.latest+1 {
		logger.Debug().Str("hash", block.ID()).Msg("block is branch")
		longest = false
	}

	logger.Debug().Uint64("blockNo", blockNo).Str("hash", block.ID()).Msg("add block to db")
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	tx := *dbtx
	if longest {
		tx.Set(latestKey, blockKey)
		tx.Set(blockKey, block.GetHash())
	}
	tx.Set(block.GetHash(), blockBytes)

	// to avoid exception, set here
	if longest {
		cdb.setLatest(blockNo)
	}

	return nil
}
func (cdb *ChainDB) getBestBlockNo() types.BlockNo {
	return cdb.latest
}
func (cdb *ChainDB) getBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	blockHash, err := cdb.getHashByNo(blockNo)
	if err != nil {
		return nil, err
	}
	//logger.Debugf("getblockbyNo No=%d Hash=%v", blockNo, EncodeB64(blockHash))
	return cdb.getBlock(blockHash)
}
func (cdb *ChainDB) getBlock(blockHash []byte) (*types.Block, error) {
	if blockHash == nil {
		return nil, fmt.Errorf("block hash invalid(nil)")
	}
	buf := types.Block{}
	err := cdb.loadData(blockHash, &buf)
	if err != nil {
		return nil, fmt.Errorf("block not found: blockHash=%v", EncodeB64(blockHash))
	}

	//logger.Debugf("getblockbyHash Hash=%v", EncodeB64(blockHash))
	return &buf, nil
}
func (cdb *ChainDB) getHashByNo(blockNo types.BlockNo) ([]byte, error) {
	blockKey := ItobU64(uint64(blockNo))
	blockHash := cdb.store.Get(blockKey)
	if blockHash == nil || len(blockHash) == 0 {
		return nil, fmt.Errorf("block not found: blockNo=%d", blockNo)
	}
	return blockHash, nil
}
func (cdb *ChainDB) getTx(txHash []byte) (*types.Tx, *types.TxIdx, error) {
	txIdx := &types.TxIdx{}

	err := cdb.loadData(txHash, txIdx)
	if err != nil {
		return nil, nil, fmt.Errorf("tx not found: txHash=%v", EncodeB64(txHash))
	}
	block, err := cdb.getBlock(txIdx.BlockHash)
	if err != nil {
		return nil, nil, fmt.Errorf("block not found: blockHash=%v", EncodeB64(txIdx.BlockHash))
	}
	txs := block.GetBody().GetTxs()
	if txIdx.Idx >= int32(len(txs)) {
		return nil, nil, fmt.Errorf("wrong tx idx: %d", txIdx.Idx)
	}
	tx := txs[txIdx.Idx]
	logger.Debug().Str("hash", EncodeB64(txHash)).Msg("getTx")

	return tx, txIdx, nil
}

type ChainTree struct {
	Tree []ChainInfo
}
type ChainInfo struct {
	Height types.BlockNo
	Hash   string
}

func (cdb *ChainDB) GetChainTree() ([]byte, error) {
	tree := make([]ChainInfo, 0)
	var i uint64
	for i = 0; i < cdb.latest; i++ {
		hash, _ := cdb.getHashByNo(i)
		tree = append(tree, ChainInfo{
			Height: i,
			Hash:   EncodeB64(hash),
		})
		logger.Info().Str("hash", EncodeB64(hash)).Msg("GetChainTree")
	}
	jsonBytes, err := json.Marshal(tree)
	if err != nil {
		logger.Info().Msg("GetChainTree failed")
	}
	return jsonBytes, nil
}
