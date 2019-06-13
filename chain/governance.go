/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"math/big"

	"github.com/aergoio/aergo/contract/enterprise"
	"github.com/aergoio/aergo/contract/name"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func executeGovernanceTx(bs *state.BlockState, txBody *types.TxBody, sender, receiver *state.V,
	blockNo types.BlockNo) ([]*types.Event, error) {

	if len(txBody.Payload) <= 0 {
		return nil, types.ErrTxFormatInvalid
	}

	governance := string(txBody.Recipient)

	scs, err := bs.StateDB.OpenContractState(receiver.AccountID(), receiver.State())
	if err != nil {
		return nil, err
	}
	var events []*types.Event
	switch governance {
	case types.AergoSystem:
		events, err = system.ExecuteSystemTx(scs, txBody, sender, receiver, blockNo)
	case types.AergoName:
		events, err = name.ExecuteNameTx(bs, scs, txBody, sender, receiver, blockNo)
	case types.AergoEnterprise:
		events, err = enterprise.ExecuteEnterpriseTx(scs, txBody, sender)
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}
	if err == nil {
		err = bs.StateDB.StageContractState(scs)
	}

	return events, err
}

// InitGenesisBPs opens system contract and put initial voting result
// it also set *State in Genesis to use statedb
func InitGenesisBPs(states *state.StateDB, genesis *types.Genesis) error {
	aid := types.ToAccountID([]byte(types.AergoSystem))
	scs, err := states.OpenContractStateAccount(aid)
	if err != nil {
		return err
	}

	voteResult := make(map[string]*big.Int)
	for _, v := range genesis.BPs {
		voteResult[v] = new(big.Int).SetUint64(0)
	}
	if err = system.InitVoteResult(scs, voteResult); err != nil {
		return err
	}

	// Set genesis.BPs to the votes-ordered BPs. This will be used later for
	// bootstrapping.
	genesis.BPs = system.BuildOrderedCandidates(voteResult)
	if err = states.StageContractState(scs); err != nil {
		return err
	}
	if err = states.Update(); err != nil {
		return err
	}
	if err = states.Commit(); err != nil {
		return err
	}

	return nil
}
