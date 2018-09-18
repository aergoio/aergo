/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"github.com/aergoio/aergo/state"

	"github.com/aergoio/aergo/types"
)

const minimum = 1000
const aergosystem = "aergo.system"

func executeGovernanceTx(sdb *state.ChainStateDB, txBody *types.TxBody, senderState *types.State, receiverState *types.State,
	blockNo types.BlockNo) error {
	governance := string(txBody.GetRecipient())

	scs, err := sdb.OpenContractState(receiverState)
	if err != nil {
		return err
	}
	switch governance {
	case aergosystem:
		/*
			TODO: need validate?
			peerID, err := peer.IDFromBytes(to)
			if err != nil {
				return err
			}
		*/
		err = executeSystemTx(txBody, senderState, scs, blockNo)
		if err == nil {
			err = sdb.CommitContractState(scs)
		}
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
	}
	return err
}

func executeSystemTx(txBody *types.TxBody, senderState *types.State,
	scs *state.ContractState, blockNo types.BlockNo) error {
	systemCmd := txBody.GetPayload()[0]
	var err error
	switch systemCmd {
	case 's':
		err = staking(txBody, senderState, scs, blockNo)
	case 'v':
		err = voting(txBody, scs, blockNo)
	case 'u':
		err = unstaking(txBody, senderState, scs, blockNo)
	}
	if err != nil {
		return err
	}
	return nil
}
