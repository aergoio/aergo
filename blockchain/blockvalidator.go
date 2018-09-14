/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"errors"
	"github.com/aergoio/aergo/types"
)

type BlockValidator struct {
	signVerifier *SignVerifier
}

var (
	ErrorBlockVerifySign = errors.New("Block verify failed, because Tx sign is invalid")
)

func NewBlockValidator() *BlockValidator {
	bv := BlockValidator{
		signVerifier: NewSignVerifier(DefaultVerifierCnt),
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
	return nil
}

func (bv *BlockValidator) ValidateBody(block *types.Block) error {
	// TxRootHash
	// check Tx sign
	txs := block.GetBody().GetTxs()
	if len(txs) == 0 {
		return nil
	}

	failed, _ := bv.signVerifier.VerifyTxs(&types.TxList{Txs: txs})

	if failed {
		logger.Error().Str("block", block.ID()).Msg("block verify failed")
		return ErrorBlockVerifySign
	}

	return nil
}
