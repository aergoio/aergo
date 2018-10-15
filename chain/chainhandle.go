/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	sha256 "github.com/minio/sha256-simd"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
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
	//logger.Debug().Uint64("blockno", blockNo).Msg("get best block")
	block := cs.cdb.bestBlock.Load().(*types.Block)

	if block == nil {
		return nil, errors.New("best block is null")
	}
	return block, nil
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

	tx, txidx, err := cs.cdb.getTx(txHash)
	if err != nil {
		return nil, nil, err
	}
	block, err := cs.cdb.getBlock(txidx.BlockHash)
	blockInMainChain, err := cs.cdb.getBlockByNo(block.Header.BlockNo)
	if !bytes.Equal(block.BlockHash(), blockInMainChain.BlockHash()) {
		return tx, nil, errors.New("tx is not in the main chain")
	}
	return tx, txidx, err
}

type chainProcessor struct {
	*ChainService
	block     *types.Block // starting block
	lastBlock *types.Block
	state     *state.BlockState
	mainChain *list.List

	add func(blk *types.Block) error
}

func newChainProcessor(block *types.Block, state *state.BlockState, cs *ChainService) (*chainProcessor, error) {
	var isMainChain bool
	var err error

	if isMainChain, err = cs.cdb.isMainChain(block); err != nil {
		return nil, err
	}

	cp := &chainProcessor{
		ChainService: cs,
		block:        block,
		state:        state,
	}

	if isMainChain {
		cp.mainChain = list.New()
		cp.add = func(blk *types.Block) error {
			if err := cp.addCommon(blk); err != nil {
				return err
			}
			// blk must be executed later if it belongs to the main chain.
			cp.mainChain.PushBack(blk)

			return nil
		}
	} else {
		cp.add = cp.addCommon
	}

	return cp, nil
}

func (cp *chainProcessor) addCommon(blk *types.Block) error {
	dbTx := cp.cdb.store.NewTx()
	defer dbTx.Discard()

	if err := cp.cdb.addBlock(&dbTx, blk); err != nil {
		return err
	}

	dbTx.Commit()

	logger.Debug().Bool("isMainChain", cp.isMain()).
		Uint64("latest", cp.cdb.latest).
		Uint64("blockNo", blk.BlockNo()).
		Str("hash", blk.ID()).
		Str("prev_hash", enc.ToString(blk.GetHeader().GetPrevBlockHash())).
		Msg("block added to the block indices")

	cp.lastBlock = blk

	return nil
}

func (cp *chainProcessor) prepare() error {
	var err error

	blk := cp.block
	for blk != nil {
		// Add blk to the corresponding block chain.
		if err := cp.add(blk); err != nil {
			return err
		}

		// Remove a block depnding on blk from the orphan cache.
		if blk, err = cp.resolveOrphan(blk); err != nil {
			return err
		}
	}

	return nil
}

func (cp *chainProcessor) isMain() bool {
	return cp.mainChain != nil
}

func (cp *chainProcessor) executeBlock(block *types.Block) error {
	err := cp.ChainService.executeBlock(cp.state, block)
	cp.state = nil

	return err
}

func (cp *chainProcessor) execute() error {
	if !cp.isMain() {
		return nil
	}
	logger.Debug().Int("blocks to execute", cp.mainChain.Len()).Msg("start to execute")

	var err error
	for e := cp.mainChain.Front(); e != nil; e = e.Next() {
		block := e.Value.(*types.Block)

		if err = cp.executeBlock(block); err != nil {
			logger.Error().Str("error", err.Error()).Str("hash", block.ID()).
				Msg("failed to execute block")

			return err
		}

		blockNo := block.BlockNo()

		var oldLatest types.BlockNo

		//SyncWithConsensus :ga
		// 	After executing MemPoolDel in the chain service, MemPoolGet must be executed on the consensus.
		// 	To do this, cdb.setLatest() must be executed after MemPoolDel.
		//	In this case, messages of mempool is synchronized in actor message queue.
		if oldLatest, err = cp.connectToChain(block); err != nil {
			return err
		}

		cp.notifyBlock(block)

		logger.Debug().
			Uint64("old latest", oldLatest).
			Uint64("new latest", blockNo).
			Str("hash", block.ID()).
			Str("prev_hash", enc.ToString(block.GetHeader().GetPrevBlockHash())).
			Msg("block executed")

	}

	return nil
}

func (cp *chainProcessor) connectToChain(block *types.Block) (types.BlockNo, error) {
	dbTx := cp.cdb.store.NewTx()
	defer dbTx.Discard()

	oldLatest := cp.cdb.connectToChain(&dbTx, block)

	if err := cp.cdb.addTxsOfBlock(&dbTx, block.GetBody().GetTxs(), block.BlockHash()); err != nil {
		return 0, err
	}

	dbTx.Commit()

	return oldLatest, nil
}

func (cp *chainProcessor) reorganize() {
	// - Reorganize if new bestblock then process Txs
	// - Add block if new bestblock then update context connect next orphan
	if cp.needReorg(cp.lastBlock) {
		err := cp.reorg(cp.lastBlock)
		if e, ok := err.(consensus.ErrorConsensus); ok {
			logger.Info().Err(e).Msg("stop reorganization")
			return
		}

		if err != nil {
			panic(err)
		}
	}
}

func (cs *ChainService) addBlock(newBlock *types.Block, usedBstate *state.BlockState, peerID peer.ID) error {
	logger.Debug().Str("hash", newBlock.ID()).Msg("add block")

	var bestBlock *types.Block
	var err error

	if bestBlock, err = cs.getBestBlock(); err != nil {
		return err
	}

	// Check consensus header validity
	if err := cs.IsBlockValid(newBlock, bestBlock); err != nil {
		return err
	}

	// handle orphan
	if cs.isOrphan(newBlock) {
		if usedBstate != nil {
			return fmt.Errorf("block received from BP can not be orphan")
		}
		err := cs.handleOrphan(newBlock, bestBlock, peerID)
		return err
	}

	cp, err := newChainProcessor(newBlock, usedBstate, cs)
	if err != nil {
		return err
	}

	if err := cp.prepare(); err != nil {
		return err
	}
	if err := cp.execute(); err != nil {
		return err
	}

	// TODO: reorganization should be done before chain execution to avoid an
	// unnecessary chain execution & rollback.
	cp.reorganize()

	logger.Info().Uint64("best", cs.cdb.getBestBlockNo()).Msg("added block successfully. ")

	return nil
}

func (cs *ChainService) CountTxsInChain() int {
	var txCount int

	blk, err := cs.getBestBlock()
	if err != nil {
		return -1
	}

	var no uint64
	for {
		no = blk.GetHeader().GetBlockNo()
		if no == 0 {
			break
		}

		txCount += len(blk.GetBody().GetTxs())

		blk, err = cs.getBlock(blk.GetHeader().GetPrevBlockHash())
		if err != nil {
			txCount = -1
			break
		}
	}

	return txCount
}

type TxExecFn func(bState *state.BlockState, tx *types.Tx) error

type blockExecutor struct {
	*state.BlockState
	sdb        *state.ChainStateDB
	execTx     TxExecFn
	txs        []*types.Tx
	commitOnly bool
}

func newBlockExecutor(cs *ChainService, bState *state.BlockState, block *types.Block) (*blockExecutor, error) {
	var exec TxExecFn

	commitOnly := false

	// The DPoS block factory excutes transactions during block generation. In
	// such a case it send block with block state so that bState != nil. On the
	// contrary, the block propagated from the network is not half-executed.
	// Hence we need a new block state and tx executor (execTx).
	if bState == nil {
		if err := cs.validator.ValidateBlock(block); err != nil {
			return nil, err
		}

		bState = state.NewBlockState(
			block.BlockID(),
			cs.sdb.OpenNewStateDB(cs.sdb.GetRoot()),
			contract.TempReceiptDb.NewTx(),
		)

		exec = NewTxExecutor(block.BlockNo(), block.GetHeader().GetTimestamp())
	} else {
		logger.Debug().Uint64("block no", block.BlockNo()).Msg("received block from block factory")
		// In this case (bState != nil), the transactions has already been
		// executed by the block factory.
		commitOnly = true
	}

	txs := block.GetBody().GetTxs()

	return &blockExecutor{
		BlockState: bState,
		sdb:        cs.sdb,
		execTx:     exec,
		txs:        txs,
		commitOnly: commitOnly,
	}, nil
}

// NewTxExecutor returns a new TxExecFn.
func NewTxExecutor(blockNo types.BlockNo, ts int64) TxExecFn {
	return func(bState *state.BlockState, tx *types.Tx) error {
		if bState == nil {
			logger.Error().Msg("bstate is nil in txexec")
			return ErrGatherChain
		}
		snapshot := bState.Snapshot()

		err := executeTx(bState, tx, blockNo, ts)
		if err != nil {
			logger.Error().Str("hash", enc.ToString(tx.GetHash())).Msg("tx failed")
			bState.Rollback(snapshot)
			return err
		}
		return nil
	}
}

func (e *blockExecutor) execute() error {
	// Receipt must be committed unconditionally.
	if !e.commitOnly {
		for _, tx := range e.txs {
			if err := e.execTx(e.BlockState, tx); err != nil {
				//FIXME maybe system error. restart or panic
				// all txs have executed successfully in BP node
				return err
			}
		}

		if err := e.Update(); err != nil {
			return err
		}
	}

	err := contract.SaveRecoveryPoint(e.BlockState)
	if err != nil {
		return err
	}

	// TODO: sync status of bstate and cdb what to do if cdb.commit fails after
	// sdb.Apply() succeeds
	err = e.commit()

	return err
}

func (e *blockExecutor) commit() error {
	e.CommitReceipt()

	if err := e.BlockState.Commit(); err != nil {
		return err
	}

	//TODO: after implementing BlockRootHash, remove statedb.lastest
	if err := e.sdb.UpdateRoot(e.BlockState); err != nil {
		return err
	}

	return nil
}

//TODO Refactoring: batch
func (cs *ChainService) executeBlock(bstate *state.BlockState, block *types.Block) error {
	ex, err := newBlockExecutor(cs, bstate, block)
	if err != nil {
		return err
	}

	// contract & state DB update is done during execution.
	if err := ex.execute(); err != nil {
		// FIXME: is that enough?
		logger.Error().Err(err).Str("hash", block.ID()).Msg("failed to execute block")

		return err
	}

	cs.RequestTo(message.MemPoolSvc, &message.MemPoolDel{
		Block: block,
	})

	cs.UpdateStatus(block)

	return nil
}

func executeTx(bs *state.BlockState, tx *types.Tx, blockNo uint64, ts int64) error {
	txBody := tx.GetBody()
	senderID := types.ToAccountID(txBody.Account)
	senderState, err := bs.GetAccountState(senderID)
	if err != nil {
		return err
	}
	recipient := txBody.Recipient
	var receiverID types.AccountID
	var createContract bool
	if len(recipient) > 0 {
		receiverID = types.ToAccountID(recipient)
	} else {
		createContract = true
		// Determine new contract address
		h := sha256.New()
		h.Write(txBody.Account)
		h.Write([]byte(strconv.FormatUint(txBody.Nonce, 10)))
		recipientHash := h.Sum(nil)                        // byte array with length 32
		recipient = append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
		receiverID = types.ToAccountID(recipient)
	}
	receiverState, err := bs.GetAccountState(receiverID)
	if err != nil {
		return err
	}

	senderChange := types.State(*senderState)
	receiverChange := types.State(*receiverState)

	switch txBody.Type {
	case types.TxType_NORMAL:
		if senderID != receiverID {
			if senderChange.Balance < txBody.Amount {
				senderChange.Balance = 0 // FIXME: reject insufficient tx.
			} else {
				senderChange.Balance = senderState.Balance - txBody.Amount
			}
			receiverChange.Balance = receiverChange.Balance + txBody.Amount
		}
		if txBody.Payload != nil {
			contractState, err := bs.OpenContractState(&receiverChange)
			if err != nil {
				return err
			}
			if createContract {
				receiverChange.SqlRecoveryPoint = 1
			}
			sqlTx, err := contract.BeginTx(receiverID, receiverChange.SqlRecoveryPoint)
			if err != nil {
				return err
			}
			err = sqlTx.Savepoint()
			if err != nil {
				return err
			}

			bcCtx := contract.NewContext(bs, &senderChange, contractState, types.EncodeAddress(txBody.GetAccount()),
				hex.EncodeToString(tx.GetHash()), blockNo, ts, "", 0,
				types.EncodeAddress(recipient), 0, nil, sqlTx.GetHandle())

			if createContract {
				err = contract.Create(contractState, txBody.Payload, recipient, tx.Hash, bcCtx, bs.ReceiptTx())
			} else {
				err = contract.Call(contractState, txBody.Payload, recipient, tx.Hash, bcCtx, bs.ReceiptTx())
			}
			if err != nil { // vm error is not propagated
				return sqlTx.RollbackToSavepoint()
			}
			err = bs.CommitContractState(contractState)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = sqlTx.Release()
			if err != nil {
				return err
			}
		}
	case types.TxType_GOVERNANCE:
		err = executeGovernanceTx(&bs.StateDB, txBody, &senderChange, &receiverChange, blockNo)
	default:
		logger.Warn().Str("tx", tx.String()).Msg("unknown type of transaction")
	}

	senderChange.Nonce = txBody.Nonce
	err = bs.PutState(senderID, &senderChange)
	if err != nil {
		return err
	}
	if senderID != receiverID {
		err = bs.PutState(receiverID, &receiverChange)
		if err != nil {
			return err
		}
	}

	return err
}

// find an orphan block which is the child of the added block
func (cs *ChainService) resolveOrphan(block *types.Block) (*types.Block, error) {
	hash := block.BlockHash()

	orphanID := types.ToBlockID(hash)
	orphan, exists := cs.op.cache[orphanID]
	if !exists {
		return nil, nil
	}

	orphanBlock := orphan.block

	if (block.GetHeader().GetBlockNo() + 1) != orphanBlock.GetHeader().GetBlockNo() {
		return nil, fmt.Errorf("invalid orphan block no (p=%d, c=%d)", block.GetHeader().GetBlockNo(),
			orphanBlock.GetHeader().GetBlockNo())
	}

	logger.Debug().Str("parentHash=", block.ID()).
		Str("orphanHash=", orphanBlock.ID()).
		Msg("connect orphan")

	cs.op.removeOrphan(orphanID)

	return orphanBlock, nil
}

func (cs *ChainService) isOrphan(block *types.Block) bool {
	prevhash := block.Header.PrevBlockHash
	_, err := cs.getBlock(prevhash)

	return err != nil
}

func (cs *ChainService) handleOrphan(block *types.Block, bestBlock *types.Block, peerID peer.ID) error {
	err := cs.addOrphan(block)
	if err != nil {
		// logging???
		logger.Debug().Str("hash", block.ID()).Msg("add Orphan Block failed")

		return err
	}
	// request missing
	orphanNo := block.GetHeader().GetBlockNo()
	bestNo := bestBlock.GetHeader().GetBlockNo()
	if block.GetHeader().GetBlockNo() < bestBlock.GetHeader().GetBlockNo()+1 {
		logger.Debug().Str("hash", block.ID()).Uint64("orphanNo", orphanNo).Uint64("bestNo", bestNo).
			Msg("skip sync with too old block")
		return nil
	}
	anchors := cs.getAnchorsFromHash(block.BlockHash())
	hashes := make([]message.BlockHash, 0)
	for _, a := range anchors {
		hashes = append(hashes, message.BlockHash(a))
	}
	cs.RequestTo(message.P2PSvc, &message.GetMissingBlocks{ToWhom: peerID, Hashes: hashes})

	return nil
}

func (cs *ChainService) addOrphan(block *types.Block) error {
	return cs.op.addOrphan(block)
}

func (cs *ChainService) handleMissing(stopHash []byte, Hashes [][]byte) ([]message.BlockHash, []types.BlockNo) {
	// 1. check endpoint is on main chain (or, return nil)
	logger.Debug().Str("hash", enc.ToString(stopHash)).Int("len", len(Hashes)).Msg("handle missing")
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

		if bytes.Equal(mainhash, mainblock.BlockHash()) {
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
	logger.Debug().Uint64("mainBlockNo", mainBlockNo).Str("mainHash", enc.ToString(mainhash)).
		Uint64("stopBlockNo", stopBlock.GetHeader().GetBlockNo()).Str("stopHash", enc.ToString(stopBlock.Hash)).
		Msg("Get hashes of missing part")
	rhashes := make([]message.BlockHash, 0, loop)
	rnos := make([]types.BlockNo, 0, loop)
	for i := uint64(0); i < loop; i++ {
		tBlock, _ := cs.getBlockByNo(types.BlockNo(mainBlockNo + i))
		rhashes = append(rhashes, message.BlockHash(tBlock.Hash))
		rnos = append(rnos, types.BlockNo(tBlock.GetHeader().GetBlockNo()))
		logger.Debug().Uint64("blockNo", tBlock.GetHeader().GetBlockNo()).Str("hash", enc.ToString(tBlock.Hash)).
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
