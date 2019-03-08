/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/gogo/protobuf/proto"
)

const (
	chainDBName       = "chain"
	genesisKey        = chainDBName + ".genesisInfo"
	genesisBalanceKey = chainDBName + ".genesisBalance"

	TxBatchMax = 10000
)

var (
	// ErrNoChainDB reports chaindb is not prepared.
	ErrNoChainDB         = fmt.Errorf("chaindb not prepared")
	ErrorLoadBestBlock   = errors.New("failed to load latest block from DB")
	ErrCantDropGenesis   = errors.New("can't drop genesis block")
	ErrTooBigResetHeight = errors.New("reset height is too big")

	latestKey      = []byte(chainDBName + ".latest")
	receiptsPrefix = []byte("r")
)

// ErrNoBlock reports there is no such a block with id (hash or block number).
type ErrNoBlock struct {
	id interface{}
}

func (e ErrNoBlock) Error() string {
	var idStr string

	switch id := e.id.(type) {
	case []byte:
		idStr = fmt.Sprintf("blockHash=%v", enc.ToString(id))
	default:
		idStr = fmt.Sprintf("blockNo=%v", id)
	}

	return fmt.Sprintf("block not found: %s", idStr)
}

type ChainDB struct {
	cc consensus.ChainConsensus

	latest    atomic.Value //types.BlockNo
	bestBlock atomic.Value // *types.Block
	//	blocks []*types.Block
	store db.DB
}

func NewChainDB() *ChainDB {
	// logger.SetLevel("debug")
	cdb := &ChainDB{
		//blocks: []*types.Block{},
	}
	cdb.latest.Store(types.BlockNo(0))

	return cdb
}

// NewTx returns a new chain DB Transaction.
func (cdb *ChainDB) NewTx() db.Transaction {
	return cdb.store.NewTx()
}

func (cdb *ChainDB) Init(dbType string, dataDir string) error {
	if cdb.store == nil {
		dbPath := common.PathMkdirAll(dataDir, chainDBName)
		cdb.store = db.NewDB(db.ImplType(dbType), dbPath)
	}

	// load data
	if err := cdb.loadChainData(); err != nil {
		return err
	}
	// // if empty then create new genesis block
	// // if cdb.getBestBlockNo() == 0 && len(cdb.blocks) == 0 {
	// blockIdx := types.BlockNoToBytes(0)
	// blockHash := cdb.store.Get(blockIdx)
	// if cdb.getBestBlockNo() == 0 && (blockHash == nil || len(blockHash) == 0) {
	// 	cdb.generateGenesisBlock(seed)
	// }
	return nil
}

// ResetBest reset best block of chain db manually remove blocks from original
// best to resetNo.
//
// *Caution*: This API is dangerous. It must be used for test blockchain only.
func (cdb *ChainDB) ResetBest(resetNo types.BlockNo) error {
	logger.Info().Uint64("reset height", resetNo).Msg("reset best block")

	best := cdb.getBestBlockNo()
	if best <= resetNo {
		logger.Error().Uint64("best", best).Uint64("reset", resetNo).Msg("too big reset height")
		return ErrTooBigResetHeight
	}

	for curNo := best; curNo > resetNo; curNo-- {
		if err := cdb.dropBlock(curNo); err != nil {
			logger.Error().Err(err).Uint64("no", curNo).Msg("failed to drop block")
			return err
		}
	}

	logger.Info().Msg("succeed to reset best block")

	return nil
}

type ErrDropBlock struct {
	pos int
}

func (err *ErrDropBlock) Error() string {
	return fmt.Sprintf("failed to drop block: pos=%d", err.pos)
}

func (cdb *ChainDB) checkBlockDropped(dropBlock *types.Block) error {
	no := dropBlock.GetHeader().GetBlockNo()
	hash := dropBlock.GetHash()
	txLen := len(dropBlock.GetBody().GetTxs())

	//check receipt
	var err error

	if txLen > 0 {
		if _, err = cdb.getReceipt(hash, no, 0); err == nil {
			return &ErrDropBlock{pos: 0}
		}
		if _, err = cdb.getReceipt(hash, no, int32(txLen-1)); err == nil {
			return &ErrDropBlock{pos: 1}
		}
	}

	//check tx
	for _, tx := range dropBlock.GetBody().GetTxs() {
		if _, _, err = cdb.getTx(tx.GetHash()); err == nil {
			return &ErrDropBlock{pos: 2}
		}
	}

	//check hash/block
	if _, err = cdb.getBlock(hash); err == nil {
		return &ErrDropBlock{pos: 3}
	}

	//check no/hash
	if _, err = cdb.getHashByNo(no); err == nil {
		return &ErrDropBlock{pos: 4}
	}

	return nil
}

func (cdb *ChainDB) Close() {
	if cdb.store != nil {
		cdb.store.Close()
	}
	return
}

// Get returns the value corresponding to key from the chain DB.
func (cdb *ChainDB) Get(key []byte) []byte {
	return cdb.store.Get(key)
}

func (cdb *ChainDB) GetBestBlock() (*types.Block, error) {
	//logger.Debug().Uint64("blockno", blockNo).Msg("get best block")
	var block *types.Block

	aopv := cdb.bestBlock.Load()

	if aopv != nil {
		block = aopv.(*types.Block)
	}

	return block, nil
}

func (cdb *ChainDB) loadChainData() error {
	latestBytes := cdb.store.Get(latestKey)
	if latestBytes == nil || len(latestBytes) == 0 {
		return nil
	}
	latestNo := types.BlockNoFromBytes(latestBytes)
	/* TODO: just checking DB
	cdb.blocks = make([]*types.Block, latestNo+1)
	for i := uint32(0); i <= latestNo; i++ {
		blockIdx := types.BlockNoToBytes(i)
		buf := types.Block{}
		err := cdb.loadData(blockIdx, &buf)
		if err != nil {
			return err
		}
		bHash := buf.CalculateBlockHash()
		if buf.Hash == nil {
			buf.Hash = bHash
		} else if !bytes.Equal(buf.Hash, bHash) {
			return fmt.Errorf("invalid Block Hash: hash=%s, check=%s",
				enc.ToString(buf.Hash), enc.ToString(bHash))
		}
		for _, v := range buf.Body.Txs {
			tHash := v.CalculateTxHash()
			if v.Hash == nil {
				v.Hash = tHash
			} else if !bytes.Equal(v.Hash, tHash) {
				return fmt.Errorf("invalid Transaction Hash: hash=%s, check=%s",
					enc.ToString(v.Hash), enc.ToString(tHash))
			}
		}
		cdb.blocks[i] = &buf
	}
	*/
	latestBlock, err := cdb.GetBlockByNo(latestNo)
	if err != nil {
		return ErrorLoadBestBlock
	}
	cdb.setLatest(latestBlock)

	// skips := true
	// for i, _ := range cdb.blocks {
	// 	if i > 3 && i+3 <= cdb.getBestBlockNo() {
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
	//logger.Debugf("  loadData: key=%d, len=%d, val=%s\n", Btoi(key), len(buf), enc.ToString(buf))
	err := proto.Unmarshal(buf, pb)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: key=%v, len=%d", key, len(buf))
	}
	//logger.Debug("  loaded: ", ToJSON(pb))
	return nil
}

func (cdb *ChainDB) addGenesisBlock(genesis *types.Genesis) error {
	block := genesis.Block()

	tx := cdb.store.NewTx()

	cdb.connectToChain(&tx, block)
	tx.Set([]byte(genesisKey), genesis.Bytes())
	if totalBalance := genesis.TotalBalance(); totalBalance != nil {
		tx.Set([]byte(genesisBalanceKey), totalBalance.Bytes())
	}

	tx.Commit()

	//logger.Info().Str("chain id", genesis.ID.ToJSON()).Str("chain id (raw)",
	//	enc.ToString(block.GetHeader().GetChainID())).Msg("Genesis Block Added")
	return nil
}

// GetGenesisInfo returns Genesis info, which is read from cdb.
func (cdb *ChainDB) GetGenesisInfo() *types.Genesis {
	if b := cdb.Get([]byte(genesisKey)); len(b) != 0 {
		genesis := types.GetGenesisFromBytes(b)
		if block, err := cdb.GetBlockByNo(0); err == nil {
			genesis.SetBlock(block)

			// genesis.ID is overwritten by the genesis block's chain
			// id. Prefer the latter since it is sort of protected the block
			// chain system (all the chaild blocks connected to the genesis
			// block).
			rawCid := genesis.Block().GetHeader().GetChainID()
			if len(rawCid) > 0 {
				cid := types.NewChainID()
				if err := cid.Read(rawCid); err == nil {
					genesis.ID = *cid
				}
			}

		}

		if v := cdb.Get([]byte(genesisBalanceKey)); len(v) != 0 {
			genesis.SetTotalBalance(v)
		}

		return genesis
	}
	return nil
}

func (cdb *ChainDB) setLatest(newBestBlock *types.Block) (oldLatest types.BlockNo) {
	oldLatest = cdb.getBestBlockNo()

	newLatest := types.BlockNo(newBestBlock.GetHeader().GetBlockNo())
	cdb.latest.Store(newLatest)
	cdb.bestBlock.Store(newBestBlock)

	logger.Debug().Uint64("old", oldLatest).Uint64("new", newLatest).Msg("update latest block")

	return
}

func (cdb *ChainDB) connectToChain(dbtx *db.Transaction, block *types.Block) (oldLatest types.BlockNo) {
	blockNo := block.GetHeader().GetBlockNo()
	blockIdx := types.BlockNoToBytes(blockNo)

	if err := cdb.addBlock(dbtx, block); err != nil {
		return 0
	}

	// Update best block hash
	(*dbtx).Set(latestKey, blockIdx)
	(*dbtx).Set(blockIdx, block.BlockHash())

	// Save the last consensus status.
	if cdb.cc != nil {
		if err := cdb.cc.Save(*dbtx); err != nil {
			logger.Error().Err(err).Msg("failed to save DPoS status")
		}
	}

	oldLatest = cdb.setLatest(block)

	logger.Debug().Str("hash", block.ID()).Msg("connect block to mainchain")

	return
}

func (cdb *ChainDB) swapChain(newBlocks []*types.Block) error {
	oldNo := cdb.getBestBlockNo()
	newNo := newBlocks[0].GetHeader().GetBlockNo()

	if oldNo >= newNo {
		logger.Error().Uint64("old", oldNo).Uint64("new", newNo).
			Msg("New chain is not longger than old chain")
		return ErrInvalidSwapChain
	}

	var blockIdx []byte
	var dbTx db.Transaction

	txCnt := 0

	dbTx = cdb.store.NewTx()
	defer dbTx.Discard()

	//make newTx because of batchsize limit of DB
	getNewTx := func(remainTxCnt int) {
		if txCnt+remainTxCnt >= TxBatchMax {
			dbTx.Commit()
			dbTx = cdb.store.NewTx()
			txCnt = 0
		}
	}

	for i := len(newBlocks) - 1; i >= 0; i-- {
		block := newBlocks[i]
		blockIdx = types.BlockNoToBytes(block.GetHeader().GetBlockNo())

		dbTx.Set(blockIdx, block.BlockHash())

		txCnt++

		getNewTx(0)
	}

	getNewTx(5)

	dbTx.Set(latestKey, blockIdx)

	// Save the last consensus status.
	cdb.cc.Save(dbTx)

	dbTx.Commit()

	cdb.setLatest(newBlocks[0])

	return nil
}

func (cdb *ChainDB) isMainChain(block *types.Block) (bool, error) {
	blockNo := block.GetHeader().GetBlockNo()
	bestNo := cdb.getBestBlockNo()
	if blockNo > 0 && blockNo != bestNo+1 {
		logger.Debug().Uint64("blkno", blockNo).Uint64("latest", bestNo).Msg("block is branch")

		return false, nil
	}

	prevHash := block.GetHeader().GetPrevBlockHash()
	latestHash, err := cdb.getHashByNo(cdb.getBestBlockNo())
	if err != nil { //need assertion
		return false, fmt.Errorf("failed to getting block hash by no(%v)", cdb.getBestBlockNo())
	}

	isMainChain := bytes.Equal(prevHash, latestHash)

	logger.Debug().Bool("isMainChain", isMainChain).Uint64("blkno", blockNo).Str("hash", block.ID()).
		Str("latest", enc.ToString(latestHash)).Str("prev", enc.ToString(prevHash)).
		Msg("check if block is in main chain")

	return isMainChain, nil
}

type txInfo struct {
	blockHash []byte
	idx       int
}

func (cdb *ChainDB) addTxsOfBlock(dbTx *db.Transaction, txs []*types.Tx, blockHash []byte) error {
	for i, txEntry := range txs {
		if err := cdb.addTx(dbTx, txEntry, blockHash, i); err != nil {
			logger.Error().Err(err).Str("hash", enc.ToString(blockHash)).Int("txidx", i).
				Msg("failed to add tx")

			return err
		}
	}

	return nil
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

	// TODO: Is it possible?
	// if blockNo != 0 && isMainChain && cdb.getBestBlockNo()+1 != blockNo {
	// 	return fmt.Errorf("failed to add block(%d,%v). blkno != latestNo(%d) + 1", blockNo,
	// 		block.BlockHash(), cdb.getBestBlockNo())
	// }
	// FIXME: blockNo 0 exception handling
	// assumption: not an orphan
	// fork can be here
	logger.Debug().Uint64("blockNo", blockNo).Str("hash", block.ID()).Msg("add block to db")
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		logger.Error().Err(err).Uint64("no", blockNo).Str("hash", block.ID()).Msg("failed to add block")
		return err
	}

	//add block
	(*dbtx).Set(block.BlockHash(), blockBytes)

	return nil
}

// drop block from DB
func (cdb *ChainDB) dropBlock(dropNo types.BlockNo) error {
	logger.Info().Uint64("no", dropNo).Msg("drop block")

	dbTx := cdb.NewTx()
	defer dbTx.Discard()

	if dropNo <= 0 {
		return ErrCantDropGenesis
	}

	dropBlock, err := cdb.GetBlockByNo(dropNo)
	if err != nil {
		return err
	}

	// remove tx mapping
	for _, tx := range dropBlock.GetBody().GetTxs() {
		cdb.deleteTx(&dbTx, tx)
	}

	// remove receipt
	cdb.deleteReceipts(&dbTx, dropBlock.BlockHash(), dropBlock.BlockNo())

	// remove (hash/block)
	dbTx.Delete(dropBlock.BlockHash())

	// remove (no/hash)
	dropIdx := types.BlockNoToBytes(dropNo)
	newLatestIdx := types.BlockNoToBytes(dropNo - 1)
	dbTx.Delete(dropIdx)

	// update latest
	dbTx.Set(latestKey, newLatestIdx)

	dbTx.Commit()

	prevBlock, err := cdb.GetBlockByNo(dropNo - 1)
	if err != nil {
		return err
	}

	cdb.setLatest(prevBlock)

	if err = cdb.checkBlockDropped(dropBlock); err != nil {
		logger.Error().Err(err).Msg("block meta is not dropped")
		return err
	}
	return nil
}

func (cdb *ChainDB) getBestBlockNo() (latestNo types.BlockNo) {
	aopv := cdb.latest.Load()
	if aopv != nil {
		latestNo = aopv.(types.BlockNo)
	} else {
		panic("ChainDB:latest is nil")
	}
	return latestNo
}

// GetBlockByNo returns the block with its block number as blockNo.
func (cdb *ChainDB) GetBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	blockHash, err := cdb.getHashByNo(blockNo)
	if err != nil {
		return nil, err
	}
	//logger.Debugf("getblockbyNo No=%d Hash=%v", blockNo, enc.ToString(blockHash))
	return cdb.getBlock(blockHash)
}

func (cdb *ChainDB) getBlock(blockHash []byte) (*types.Block, error) {
	if blockHash == nil {
		return nil, fmt.Errorf("block hash invalid(nil)")
	}
	buf := types.Block{}
	err := cdb.loadData(blockHash, &buf)
	if err != nil || !bytes.Equal(buf.Hash, blockHash) {
		return nil, &ErrNoBlock{id: blockHash}
	}

	//logger.Debugf("getblockbyHash Hash=%v", enc.ToString(blockHash))
	return &buf, nil
}

func (cdb *ChainDB) getHashByNo(blockNo types.BlockNo) ([]byte, error) {
	blockIdx := types.BlockNoToBytes(blockNo)
	if cdb.store == nil {
		return nil, ErrNoChainDB
	}
	blockHash := cdb.store.Get(blockIdx)
	if len(blockHash) == 0 {
		return nil, &ErrNoBlock{id: blockNo}
	}
	return blockHash, nil
}

func (cdb *ChainDB) getTx(txHash []byte) (*types.Tx, *types.TxIdx, error) {
	txIdx := &types.TxIdx{}

	err := cdb.loadData(txHash, txIdx)
	if err != nil {
		return nil, nil, fmt.Errorf("tx not found: txHash=%v", enc.ToString(txHash))
	}
	block, err := cdb.getBlock(txIdx.BlockHash)
	if err != nil {
		return nil, nil, &ErrNoBlock{txIdx.BlockHash}
	}
	txs := block.GetBody().GetTxs()
	if txIdx.Idx >= int32(len(txs)) {
		return nil, nil, fmt.Errorf("wrong tx idx: %d", txIdx.Idx)
	}
	tx := txs[txIdx.Idx]
	logger.Debug().Str("hash", enc.ToString(txHash)).Msg("getTx")

	return tx, txIdx, nil
}

func (cdb *ChainDB) getReceipt(blockHash []byte, blockNo types.BlockNo, idx int32) (*types.Receipt, error) {
	storedReceipts, err := cdb.getReceipts(blockHash, blockNo)
	if err != nil {
		return nil, err
	}
	receipts := storedReceipts.Get()

	if idx < 0 || idx > int32(len(receipts)) {
		return nil, fmt.Errorf("cannot find a receipt: invalid index (%d)", idx)
	}
	r := receipts[idx]
	r.SetMemoryInfo(blockHash, blockNo, idx)
	return receipts[idx], nil
}

func (cdb *ChainDB) getReceipts(blockHash []byte, blockNo types.BlockNo) (*types.Receipts, error) {
	data := cdb.store.Get(receiptsKey(blockHash, blockNo))
	if len(data) == 0 {
		return nil, errors.New("cannot find a receipt")
	}
	var b bytes.Buffer
	b.Write(data)
	var receipts types.Receipts
	decoder := gob.NewDecoder(&b)
	decoder.Decode(&receipts)

	return &receipts, nil
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
	for i = 0; i < cdb.getBestBlockNo(); i++ {
		hash, _ := cdb.getHashByNo(i)
		tree = append(tree, ChainInfo{
			Height: i,
			Hash:   enc.ToString(hash),
		})
		logger.Info().Str("hash", enc.ToString(hash)).Msg("GetChainTree")
	}
	jsonBytes, err := json.Marshal(tree)
	if err != nil {
		logger.Info().Msg("GetChainTree failed")
	}
	return jsonBytes, nil
}

func (cdb *ChainDB) writeReceipts(blockHash []byte, blockNo types.BlockNo, receipts *types.Receipts) {
	dbTx := cdb.store.NewTx()
	defer dbTx.Discard()

	var val bytes.Buffer
	gob := gob.NewEncoder(&val)
	gob.Encode(receipts)

	dbTx.Set(receiptsKey(blockHash, blockNo), val.Bytes())

	dbTx.Commit()
}

func (cdb *ChainDB) deleteReceipts(dbTx *db.Transaction, blockHash []byte, blockNo types.BlockNo) {
	(*dbTx).Delete(receiptsKey(blockHash, blockNo))
}

func receiptsKey(blockHash []byte, blockNo types.BlockNo) []byte {
	var key bytes.Buffer
	key.Write(receiptsPrefix)
	key.Write(blockHash)
	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l[:], blockNo)
	key.Write(l)
	return key.Bytes()
}
