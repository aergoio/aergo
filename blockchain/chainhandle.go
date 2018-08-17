/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
)

func (cs *ChainService) getBestBlockNo() types.BlockNo {
	return cs.cdb.getBestBlockNo()
}
func (cs *ChainService) GetBestBlock() (*types.Block, error) {
	return cs.getBestBlock()
}
func (cs *ChainService) getBestBlock() (*types.Block, error) {
	blockNo := cs.cdb.getBestBlockNo()
	return cs.cdb.getBlockByNo(blockNo)
}

func (cs *ChainService) getBlockByNo(blockNo types.BlockNo) (*types.Block, error) {
	return cs.cdb.getBlockByNo(blockNo)
}

func (cs *ChainService) getBlock(blockHash []byte) (*types.Block, error) {
	return cs.cdb.getBlock(blockHash)
}

func (cs *ChainService) getHashByNo(blockNo types.BlockNo) ([]byte, error) {
	return cs.cdb.getHashByNo(blockNo)
}

func (cs *ChainService) getTx(txHash []byte) (*types.Tx, *types.TxIdx, error) {
	return cs.cdb.getTx(txHash)
}

func (cs *ChainService) addBlock(nblock *types.Block, peerID peer.ID) error {
	logger.Debug().Str("hash", nblock.ID()).Msg("add block")
	if cs.ChainInfo != nil {
		// Check block validity by calling the corresponding interface
		// implemented in a Consensus module.
		if err := cs.IsBlockValid(nblock); err != nil {
			logger.Error().Err(err).Msg("failed to add block. block is invalid.")
			return err
		}
	}

	// handle orphan
	if cs.isOrphan(nblock) {
		err := cs.handleOrphan(nblock, peerID)
		return err
	}

	// connect orphans
	block := nblock

	for block != nil {
		dbtx := cs.cdb.store.NewTx(true)

		/* reorgnize
		   if new bestblock then process Txs
		   add block
		   if new bestblock then update context
		   connect next orphan
		*/
		if cs.needReorg(block) {
			cs.reorg(&dbtx, block)
		}

		if cs.cdb.isNewBestBlock(block) {
			if err := cs.processTxsAndState(&dbtx, block); err != nil {
				return err
			}
		}

		err := cs.cdb.addBlock(&dbtx, block)
		if err != nil {
			logger.Error().Err(err).Str("hash", block.ID()).Msg("failed to add block")
			return err
		}

		dbtx.Commit()

		logger.Info().Int("processed_txn", len(block.GetBody().GetTxs())).
			Uint64("blockNo", block.GetHeader().GetBlockNo()).
			Str("hash", block.ID()).
			Str("prev_hash", EncodeB64(block.GetHeader().GetPrevBlockHash())).Msg("block added")
		//return cs.mpool.Removes(block.GetBody().GetTxs()...)
		cs.Hub().Request(message.MemPoolSvc, &message.MemPoolDel{
			// FIXME: remove legacy
			BlockNo: block.GetHeader().GetBlockNo(),
			Txs:     block.GetBody().GetTxs(),
		}, cs)

		if block, err = cs.connectOrphan(block); err != nil {
			return err
		}
	}

	cs.notifyBlock(nblock)

	return nil
}

func (cs *ChainService) processTxsAndState(dbtx *db.Transaction, block *types.Block) error {
	blockHash := types.ToBlockID(block.GetHash())
	prevHash := types.ToBlockID(block.GetHeader().GetPrevBlockHash())

	bstate := state.NewBlockState(block.Header.BlockNo, blockHash, prevHash)
	txs := block.GetBody().GetTxs()

	logger.Debug().Uint64("blockNo", block.GetHeader().GetBlockNo()).Str("hash", block.ID()).Msg("process txs and update state")

	for i, tx := range txs {
		err := cs.processTx(dbtx, bstate, tx, block.Hash, i)
		if err != nil {
			logger.Error().Err(err).Str("hash", block.ID()).Int("txidx", i).Msg("failed to process tx")
			return err
		}
	}
	err := cs.sdb.Apply(bstate)
	if err != nil {
		// FIXME: is that enough?
		logger.Error().Err(err).Str("hash", block.ID()).Msg("failed to apply state")
		return err
	}

	return nil
}

func (cs *ChainService) rollbackBlock(dbtx *db.Transaction, block *types.Block) error {
	txs := block.GetBody().GetTxs()

	blockNo := block.GetHeader().GetBlockNo()

	logger.Debug().Uint64("blockNo", blockNo).Str("hash", block.ID()).
		Msg("rollback txs of block")

	if blockNo == 0 {
		return fmt.Errorf("rollback target no can not be 0")
	}

	if blockNo != cs.cdb.latest {
		return fmt.Errorf("rollback target is not latest.(target:%v,latest:%v)",
			blockNo, cs.cdb.latest)
	}

	for _, tx := range txs {
		cs.cdb.deleteTx(dbtx, tx)
	}

	//update best block
	cs.cdb.setLatest(blockNo - 1)

	return nil
}

func (cs *ChainService) processTx(dbtx *db.Transaction, bs *state.BlockState, tx *types.Tx, blockHash []byte, idx int) error {
	txBody := tx.GetBody()
	senderID := types.ToAccountID(txBody.Account)
	senderState, err := cs.sdb.GetAccountClone(bs, senderID)
	if err != nil {
		return err
	}
	receiverID := types.ToAccountID(txBody.Recipient)
	receiverState, err := cs.sdb.GetAccountClone(bs, receiverID)
	if err != nil {
		return err
	}

	senderChange := types.Clone(*senderState).(types.State)
	receiverChange := types.Clone(*receiverState).(types.State)
	if senderID != receiverID {
		if senderChange.Balance < txBody.Amount {
			senderChange.Balance = 0 // FIXME: reject insufficient tx.
		} else {
			senderChange.Balance = senderState.Balance - txBody.Amount
		}
		receiverChange.Balance = receiverChange.Balance + txBody.Amount
		bs.PutAccount(receiverID, receiverState, &receiverChange)
	}
	senderChange.Nonce = txBody.Nonce
	bs.PutAccount(senderID, senderState, &senderChange)

	// logger.Infof("  - amount(%d), sender(%s, %s), recipient(%s, %s)",
	// 	txBody.Amount, senderID, senderState.ToString(),
	// 	receiverID, receiverState.ToString())

	err = cs.cdb.addTx(dbtx, tx, blockHash, idx)
	return err
}

// find an orphan block which is the child of the added block
func (cs *ChainService) connectOrphan(block *types.Block) (*types.Block, error) {
	hash := block.GetHash()
	for key, orphan := range cs.op.cache {
		phash := orphan.block.GetHeader().GetPrevBlockHash()
		orphanBlock := orphan.block

		if bytes.Equal(phash, hash) {
			if (block.GetHeader().GetBlockNo() + 1) != orphanBlock.GetHeader().GetBlockNo() {
				return nil, fmt.Errorf("invalid orphan block no (p=%d, c=%d)", block.GetHeader().GetBlockNo(),
					orphanBlock.GetHeader().GetBlockNo())
			}

			logger.Debug().Str("parentHash=", block.ID()).
				Uint64("blockNo", block.GetHeader().GetBlockNo()).
				Str("childHash=", orphanBlock.ID()).
				Uint64("blockNo", orphanBlock.GetHeader().GetBlockNo()).
				Msg("Connect orphan")
			cs.op.removeOrphan(key)
			return orphanBlock, nil
		}
	}
	return nil, nil
}

func (cs *ChainService) isOrphan(block *types.Block) bool {
	prevhash := block.Header.PrevBlockHash
	_, err := cs.getBlock(prevhash)

	return err != nil
}

func (cs *ChainService) handleOrphan(block *types.Block, peerID peer.ID) error {
	err := cs.addOrphan(block)
	if err != nil {
		// logging???
		logger.Debug().Str("hash", block.ID()).Msg("add Orphan Block failed")

		return err
	}
	// request missing
	anchors := cs.getAnchorsFromHash(block.Hash)
	hashes := make([]message.BlockHash, 0)
	for _, a := range anchors {
		hashes = append(hashes, message.BlockHash(a))
		logger.Debug().Str("hash", EncodeB64(a)).Msg("request block for orphan handle")
	}
	cs.Hub().Request(message.P2PSvc, &message.GetMissingBlocks{ToWhom: peerID, Hashes: hashes}, cs)

	return nil
}

func (cs *ChainService) addOrphan(block *types.Block) error {
	return cs.op.addOrphan(block)
}

func (cs *ChainService) handleMissing(stopHash []byte, Hashes [][]byte) ([]message.BlockHash, []types.BlockNo) {
	// 1. check endpoint is on main chain (or, return nil)
	logger.Debug().Str("hash", EncodeB64(stopHash)).Int("len", len(Hashes)).Msg("handle missing")
	var stopBlock *types.Block
	var err error
	if stopHash == nil {
		stopBlock, err = cs.getBestBlock()
	} else {
		stopBlock, err = cs.cdb.getBlock(stopHash)
	}
	if err != nil {
		return nil, nil
	}

	var mainhash []byte
	var mainblock *types.Block
	// 2. get the highest block of Hashes hash on main chain
	for _, hash := range Hashes {
		// need to be short
		mainblock, err = cs.cdb.getBlock(hash)
		if err != nil {
			continue
		}
		// get main hash with same block height
		mainhash, err = cs.cdb.getHashByNo(
			types.BlockNo(mainblock.GetHeader().GetBlockNo()))
		if err != nil {
			continue
		}

		if bytes.Equal(mainhash, mainblock.Hash) {
			break
		}
		mainblock = nil
	}

	// TODO: handle the case that can't find the hash in main chain
	if mainblock == nil {
		logger.Debug().Msg("Can't search same ancestor")
		return nil, nil
	}

	// 3. collect missing parts and reply them
	mainBlockNo := mainblock.GetHeader().GetBlockNo()
	var loop = stopBlock.GetHeader().GetBlockNo() - mainBlockNo
	logger.Debug().Uint64("mainBlockNo", mainBlockNo).Str("mainHash", EncodeB64(mainhash)).
		Uint64("stopBlockNo", stopBlock.GetHeader().GetBlockNo()).Str("stopHash", EncodeB64(stopBlock.Hash)).
		Msg("Get hashes of missing part")
	rhashes := make([]message.BlockHash, 0, loop)
	rnos := make([]types.BlockNo, 0, loop)
	for i := uint64(0); i < loop; i++ {
		tBlock, _ := cs.getBlockByNo(types.BlockNo(mainBlockNo + i))
		rhashes = append(rhashes, message.BlockHash(tBlock.Hash))
		rnos = append(rnos, types.BlockNo(tBlock.GetHeader().GetBlockNo()))
		logger.Debug().Uint64("blockNo", tBlock.GetHeader().GetBlockNo()).Str("hash", EncodeB64(tBlock.Hash)).
			Msg("append block for replying missing tree")
	}

	return rhashes, rnos
}

func (cs *ChainService) checkBlockHandshake(peerID peer.ID, bestHeight uint64, bestHash []byte) {
	myBestBlock, err := cs.getBestBlock()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get best block")
		return
	}
	sameBestHash := bytes.Equal(myBestBlock.Hash, bestHash)
	if sameBestHash {
		// two node has exact best block.
		// TODO: myBestBlock.GetHeader().BlockNo == bestHeight
		logger.Debug().Str("peer", peerID.Pretty()).Msg("peer is in sync status")
	} else if !sameBestHash && myBestBlock.GetHeader().BlockNo < bestHeight {
		cs.ChainSync(peerID)
	}

	return
}
