/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"errors"
	"github.com/aergoio/aergo/types"
)

type BlockVerifier struct {
	signVerifier *SignVerifier
}

var (
	ErrorBlockVerifySign = errors.New("Block verify failed, because Tx sign is invalid")
)

func NewBlockVerifier() *BlockVerifier {
	bv := BlockVerifier{
		signVerifier: NewSignVerifier(DefaultVerifierCnt),
	}

	logger.Debug().Msg("started signverifier")
	return &bv
}

func (bv *BlockVerifier) Stop() {
	bv.signVerifier.Stop()
}

func (bv *BlockVerifier) VerifyBlock(block *types.Block) error {
	if err := bv.verifyHeader(block.GetHeader()); err != nil {
		return err
	}

	if err := bv.verifyBody(block); err != nil {
		return err
	}
	return nil
}

func (bv *BlockVerifier) verifyHeader(header *types.BlockHeader) error {
	// TODO : more field?
	// Block, State not exsit
	//	MaxBlockSize
	//	MaxHeaderSize
	//	ChainVersion
	//	StateRootHash
	return nil
}

func (bv *BlockVerifier) verifyBody(block *types.Block) error {
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
