/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"errors"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const minimum = 1000
const aergobp = "aergo.bp"

func (cs *ChainService) processGovernanceTx(bs *state.BlockState, txBody *types.TxBody, block *types.Block) error {
	if txBody.Amount < minimum {
		return errors.New("too small amount to influence")
	}
	governance := string(txBody.GetRecipient())

	scs, err := cs.sdb.OpenContractStateAccount(types.ToAccountID(txBody.GetRecipient()))
	if err != nil {
		return err
	}
	switch governance {
	case aergobp:
		err = cs.processVoteTx(bs, scs, txBody, block)
		if err == nil {
			err = cs.sdb.CommitContractState(scs)
		}
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
	}
	return err
}

func (cs *ChainService) loadGovernace() error {
	scs, err := cs.sdb.OpenContractStateAccount(types.ToAccountID([]byte(aergobp)))
	if err != nil {
		return err
	}
	return cs.loadVotes(scs)
}
