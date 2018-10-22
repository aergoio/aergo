/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"github.com/aergoio/aergo/chain/system"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func executeGovernanceTx(states *state.StateDB, txBody *types.TxBody, senderState *types.State, receiverState *types.State,
	blockNo types.BlockNo) error {
	governance := string(txBody.GetRecipient())

	scs, err := states.OpenContractState(receiverState)
	if err != nil {
		return err
	}
	switch governance {
	case types.AergoSystem:
		err = system.ExecuteSystemTx(txBody, senderState, scs, blockNo)
		if err == nil {
			err = states.CommitContractState(scs)
		}
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
		err = types.ErrInvalidRecipient
	}
	return err
}

// InitGenesisBPs opens system contract and put initial voting result
// it also set *State in Genesis to use statedb
func InitGenesisBPs(states *state.StateDB, genesis *types.Genesis) error {

	if len(genesis.BPIds) == 0 {
		return nil
	}
	aid := types.ToAccountID([]byte(types.AergoSystem))
	scs, err := states.OpenContractStateAccount(aid)
	if err != nil {
		return err
	}

	voteResult := make(map[string]uint64)
	for _, v := range genesis.BPIds {
		voteResult[v] = uint64(0)
	}
	if err = system.InitVoteResult(scs, &voteResult); err != nil {
		return err
	}
	if err = states.CommitContractState(scs); err != nil {
		return err
	}
	genesis.VoteState = scs.State
	return nil
}
