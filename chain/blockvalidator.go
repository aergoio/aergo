/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type BlockValidator struct {
	signVerifier *SignVerifier
	sdb          *state.ChainStateDB
	isNeedWait   bool
	verbose      bool
}

var (
	ErrorBlockVerifySign           = errors.New("Block verify failed")
	ErrorBlockVerifyTxRoot         = errors.New("Block verify failed, because Tx root hash is invalid")
	ErrorBlockVerifyExistStateRoot = errors.New("Block verify failed, because state root hash is already exist")
	ErrorBlockVerifyStateRoot      = errors.New("Block verify failed, because state root hash is not equal")
	ErrorBlockVerifyReceiptRoot    = errors.New("Block verify failed, because receipt root hash is not equal")
)

func NewBlockValidator(comm component.IComponentRequester, sdb *state.ChainStateDB, verbose bool) *BlockValidator {
	bv := BlockValidator{
		signVerifier: NewSignVerifier(comm, sdb, VerifierCount, dfltUseMempool),
		sdb:          sdb,
		verbose:      verbose,
	}

	logger.Info().Bool("verbose", bv.verbose).Msg("started signVerifier")
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
	// Block, State not exist
	//	MaxBlockSize
	//	MaxHeaderSize
	//	ChainVersion
	//	StateRootHash
	if bv.sdb.IsExistState(header.GetBlocksRootHash()) {
		return ErrorBlockVerifyExistStateRoot
	}

	return nil
}

type validateReport struct {
	name   string
	pass   bool
	src    []byte
	target []byte
}

func (t validateReport) toString() string {
	var (
		result string
		msgStr string
	)

	if t.pass {
		result = "pass"
	} else {
		result = "failed"
	}

	msgStr = fmt.Sprintf("%s : %s. block= %s, computed=%s", t.name, result, base58.Encode(t.src), base58.Encode(t.target))

	return msgStr
}

func (bv *BlockValidator) ValidateBody(block *types.Block) error {
	txs := block.GetBody().GetTxs()

	// TxRootHash
	logger.Debug().Int("Txlen", len(txs)).Str("TxRoot", base58.Encode(block.GetHeader().GetTxsRootHash())).
		Msg("tx root verify")

	hdrRootHash := block.GetHeader().GetTxsRootHash()
	computeTxRootHash := types.CalculateTxsRootHash(txs)

	ret := bytes.Equal(hdrRootHash, computeTxRootHash)
	if bv.verbose {
		bv.report(validateReport{name: "Verify tx merkle root", pass: ret, src: hdrRootHash, target: computeTxRootHash})
	}

	if !ret {
		logger.Error().Str("block", block.ID()).
			Str("txroot", base58.Encode(hdrRootHash)).
			Str("compute txroot", base58.Encode(computeTxRootHash)).
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

func (bv *BlockValidator) report(report validateReport) {
	logger.Info().Msg(report.toString())
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

func (bv *BlockValidator) ValidatePost(sdbRoot []byte, evmRoot []byte, receipts *types.Receipts, block *types.Block) error {
	hdrRoot := block.GetHeader().GetBlocksRootHash()

	ret := bytes.Equal(hdrRoot, sdbRoot)

	if bv.verbose {
		bv.report(validateReport{name: "Verify block state root", pass: ret, src: hdrRoot, target: sdbRoot})
	}
	if !ret {
		logger.Error().Str("block", block.ID()).
			Str("hdrroot", base58.Encode(hdrRoot)).
			Str("sdbroot", base58.Encode(sdbRoot)).
			Msg("block root hash validation failed")
		return ErrorBlockVerifyStateRoot
	}

	logger.Debug().Str("sdbroot", base58.Encode(sdbRoot)).
		Msg("block root hash validation succeed")

	hdrRoot = block.GetHeader().ReceiptsRootHash
	receiptsRoot := receipts.MerkleRoot()
	ret = bytes.Equal(hdrRoot, receiptsRoot)

	if bv.verbose {
		bv.report(validateReport{name: "Verify receipt merkle root", pass: ret, src: hdrRoot, target: receiptsRoot})
	} else if !ret {
		logger.Error().Str("block", block.ID()).
			Str("hdrroot", base58.Encode(hdrRoot)).
			Str("receipts_root", base58.Encode(receiptsRoot)).
			Msg("receipts root hash validation failed")
		return ErrorBlockVerifyReceiptRoot
	}
	logger.Debug().Str("receipts_root", base58.Encode(receiptsRoot)).
		Msg("receipt root hash validation succeed")

	return nil
}
