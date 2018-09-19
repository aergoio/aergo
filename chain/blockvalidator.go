/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"errors"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

type BlockValidator struct {
	signVerifier *SignVerifier
	sdb          *state.ChainStateDB
}

var (
	ErrorBlockVerifySign      = errors.New("Block verify failed, because Tx sign is invalid")
	ErrorBlockVerifyTxRoot    = errors.New("Block verify failed, because Tx root hash is invaild")
	ErrorBlockVerifyStateRoot = errors.New("Block verify failed, because state root hash is not equal")
)

func NewBlockValidator(sdb *state.ChainStateDB) *BlockValidator {
	bv := BlockValidator{
		signVerifier: NewSignVerifier(DefaultVerifierCnt),
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
		return ErrorBlockVerifyStateRoot
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

	failed, _ := bv.signVerifier.VerifyTxs(&types.TxList{Txs: txs})

	if failed {
		logger.Error().Str("block", block.ID()).Msg("sign of txs validation failed")
		return ErrorBlockVerifySign
	}

	return nil
}
