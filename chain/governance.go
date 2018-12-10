/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo/contract/name"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func executeGovernanceTx(bs *state.BlockState, txBody *types.TxBody, sender, receiver *state.V,
	blockNo types.BlockNo) error {

	if len(txBody.Payload) <= 0 {
		return types.ErrTxFormatInvalid
	}

	governance := string(txBody.Recipient)
	if governance != types.AergoSystem && governance != types.AergoName {
		return errors.New("receive unknown recipient")
	}

	scs, err := bs.StateDB.OpenContractState(receiver.AccountID(), receiver.State())
	if err != nil {
		return err
	}
	switch governance {
	case types.AergoSystem:
		err = system.ExecuteSystemTx(scs, txBody, sender.State(), blockNo)
	case types.AergoName:
		err = name.ExecuteNameTx(scs, txBody)
	default:
		logger.Warn().Str("governance", governance).Msg("receive unknown recipient")
		err = types.ErrTxInvalidRecipient
	}
	if err == nil {
		err = bs.StateDB.StageContractState(scs)
	}

	return err
}

// InitGenesisBPs opens system contract and put initial voting result
// it also set *State in Genesis to use statedb
func InitGenesisBPs(states *state.StateDB, bps []string) error {
	aid := types.ToAccountID([]byte(types.AergoSystem))
	scs, err := states.OpenContractStateAccount(aid)
	if err != nil {
		return err
	}

	voteResult := make(map[string]*big.Int)
	for _, v := range bps {
		voteResult[v] = new(big.Int).SetUint64(0)
	}
	if err = system.InitVoteResult(scs, &voteResult); err != nil {
		return err
	}
	if err = states.StageContractState(scs); err != nil {
		return err
	}
	if err = states.Update(); err != nil {
		return err
	}

	return nil
}
