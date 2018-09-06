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

func executeGovernanceTx(sdb *state.ChainStateDB, txBody *types.TxBody, senderState *types.State, receiverState *types.State, block *types.Block) error {
	if txBody.Amount < minimum {
		return errors.New("too small amount to influence")
	}
	governance := string(txBody.GetRecipient())

	scs, err := sdb.OpenContractState(receiverState)
	if err != nil {
		return err
	}
	switch governance {
	case aergobp:
		/*
			TODO: need validate?
			peerID, err := peer.IDFromBytes(to)
			if err != nil {
				return err
			}
		*/
		err = executeVoteTx(txBody, senderState, receiverState, scs, block)
		if err == nil {
			err = sdb.CommitContractState(scs)
		}
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
	}
	return err
}
