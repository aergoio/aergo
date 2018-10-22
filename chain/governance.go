/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func executeGovernanceTx(states *state.StateDB, txBody *types.TxBody, sender, receiver *state.V,
	blockNo types.BlockNo) error {

	if len(txBody.Payload) > 0 {
		return types.ErrTxFormatInvalid
	}

	governance := string(txBody.Recipient)
	if governance != types.AergoSystem {
		return errors.New("receive unknown recipient")
	}

	scs, err := states.OpenContractState(receiver.State())
	if err != nil {
		return err
	}

	/*
		TODO: need validate?
		peerID, err := peer.IDFromBytes(to)
		if err != nil {
			return err
		}
	*/
	err = executeSystemTx(txBody, sender.State(), scs, blockNo)
	if err != nil {
		return err
	}

	return states.CommitContractState(scs)
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
	if err = syncVoteResult(scs, &voteResult); err != nil {
		return err
	}
	if err = states.CommitContractState(scs); err != nil {
		return err
	}
	genesis.VoteState = scs.State
	return nil
}
