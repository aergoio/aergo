/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"errors"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

type BlockValidator struct {
	signVerifier *SignVerifier
	sdb          *state.ChainStateDB
	isNeedWait   bool
}

var (
	ErrorBlockVerifySign           = errors.New("Block verify failed")
	ErrorBlockVerifyTxRoot         = errors.New("Block verify failed, because Tx root hash is invaild")
	ErrorBlockVerifyExistStateRoot = errors.New("Block verify failed, because state root hash is already exist")
	ErrorBlockVerifyStateRoot      = errors.New("Block verify failed, because state root hash is not equal")
	ErrorBlockVerifyReceiptRoot    = errors.New("Block verify failed, because receipt root hash is not equal")
)

func NewBlockValidator(comm component.IComponentRequester, sdb *state.ChainStateDB) *BlockValidator {
	bv := BlockValidator{
		signVerifier: NewSignVerifier(comm, sdb, VerifierCount, dfltUseMempool),
		sdb:          sdb,
	}

	logger.Debug().Msg("started signverifier")
	return &bv
}

func (bv *BlockValidator) Stop() {
	bv.signVerifier.Stop()
}

func (bv *BlockValidator) ValidateBlock(block *types.Block) error {
	if err := bv.ValidateHeader(block.GetHeader()); err != nil {
		return err
	}

	if err := bv.ValidateBody(block); err != nil {
		return err
	}
	return nil
}

func (bv *BlockValidator) ValidateHeader(header *types.BlockHeader) error {
	// TODO : more field?
	// Block, State not exsit
	//	MaxBlockSize
	//	MaxHeaderSize
	//	ChainVersion
	//	StateRootHash
	if bv.sdb.IsExistState(header.GetBlocksRootHash()) {
		return ErrorBlockVerifyExistStateRoot
	}

	return nil
}

func (bv *BlockValidator) ValidateBody(block *types.Block) error {
	txs := block.GetBody().GetTxs()

	// TxRootHash
	logger.Debug().Int("Txlen", len(txs)).Str("TxRoot", enc.ToString(block.GetHeader().GetTxsRootHash())).
		Msg("tx root verify")

	computeTxRootHash := types.CalculateTxsRootHash(txs)
	if bytes.Equal(block.GetHeader().GetTxsRootHash(), computeTxRootHash) == false {
		logger.Error().Str("block", block.ID()).
			Str("txroot", enc.ToString(block.GetHeader().GetTxsRootHash())).
			Str("compute txroot", enc.ToString(computeTxRootHash)).
			Msg("tx root validation failed")
		return ErrorBlockVerifyTxRoot
	}

	// check Tx sign
	if len(txs) == 0 {
		return nil
	}

	bv.signVerifier.RequestVerifyTxs(&types.TxList{Txs: txs})
	bv.isNeedWait = true

	return nil
}

func (bv *BlockValidator) WaitVerifyDone() error {
	logger.Debug().Bool("need", bv.isNeedWait).Msg("wait to verify tx")

	if !bv.isNeedWait {
		return nil
	}
	bv.isNeedWait = false

	if failed, _ := bv.signVerifier.WaitDone(); failed == true {
		logger.Error().Msg("sign of txs validation failed")
		return ErrorBlockVerifySign
	}

	return nil
}

func (bv *BlockValidator) ValidatePost(sdbRoot []byte, receipts *types.Receipts, block *types.Block) error {
	hdrRoot := block.GetHeader().GetBlocksRootHash()

	if !bytes.Equal(hdrRoot, sdbRoot) {
		logger.Error().Str("block", block.ID()).
			Str("hdrroot", enc.ToString(hdrRoot)).
			Str("sdbroot", enc.ToString(sdbRoot)).
			Msg("block root hash validation failed")
		return ErrorBlockVerifyStateRoot
	}

	logger.Debug().Str("block", block.ID()).
		Str("hdrroot", enc.ToString(hdrRoot)).
		Str("sdbroot", enc.ToString(sdbRoot)).
		Msg("block root hash validation succeed")

	hdrRoot = block.GetHeader().ReceiptsRootHash
	receiptsRoot := receipts.MerkleRoot()
	if !bytes.Equal(hdrRoot, receiptsRoot) {
		logger.Error().Str("block", block.ID()).
			Str("hdrroot", enc.ToString(hdrRoot)).
			Str("receipts_root", enc.ToString(receiptsRoot)).
			Msg("receipts root hash validation failed")
		return ErrorBlockVerifyReceiptRoot
	}
	logger.Debug().Str("block", block.ID()).
		Str("hdrroot", enc.ToString(hdrRoot)).
		Str("receipts_root", enc.ToString(receiptsRoot)).
		Msg("receipt root hash validation succeed")

	return nil
}
