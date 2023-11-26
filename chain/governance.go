/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"math/big"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/contract/enterprise"
	"github.com/aergoio/aergo/v2/contract/name"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

func executeGovernanceTx(ccc consensus.ChainConsensusCluster, bs *state.BlockState, txBody *types.TxBody, sender, receiver *state.AccountState,
	blockInfo *types.BlockHeaderInfo) ([]*types.Event, error) {

	if len(txBody.Payload) <= 0 {
		return nil, types.ErrTxFormatInvalid
	}

	governance := string(txBody.Recipient)

	scs, err := state.OpenContractState(receiver.AccountID(), receiver.State(), bs.StateDB)
	if err != nil {
		return nil, err
	}
	blockNo := blockInfo.No
	var events []*types.Event
	switch governance {
	case types.AergoSystem:
		events, err = system.ExecuteSystemTx(scs, txBody, sender, receiver, blockInfo)
	case types.AergoName:
		events, err = name.ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	case types.AergoEnterprise:
		events, err = enterprise.ExecuteEnterpriseTx(bs, ccc, scs, txBody, sender, receiver, blockNo)
		if err != nil {
			err = contract.NewGovEntErr(err)
		}
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}
	if err == nil {
		err = state.StageContractState(scs, bs.StateDB)
	}

	return events, err
}

// InitGenesisBPs opens system contract and put initial voting result
// it also set *State in Genesis to use statedb
func InitGenesisBPs(states *state.StateDB, genesis *types.Genesis) error {
	scs, err := state.GetSystemAccountState(states)
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
	if err = state.StageContractState(scs, states); err != nil {
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
