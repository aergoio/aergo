package system

import (
	"math"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const FutureBlockNo = math.MaxUint64

func ExecuteSystemTx(txBody *types.TxBody, senderState *types.State,
	scs *state.ContractState, blockNo types.BlockNo) error {

	systemCmd := txBody.GetPayload()[0]
	var err error

	err = ValidateSystemTx(txBody, scs, blockNo)
	if err != nil {
		return err
	}

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

type OldVotingState struct {
	staking    uint64
	voting     uint64
	candidates []byte
	when       uint64
}

func ValidateSystemTx(txBody *types.TxBody, scs *state.ContractState, blockNo uint64) error {
	if txBody.GetAmount() < types.StakingMinimum {
		return types.ErrTooSmallAmount
	}
	if len(txBody.Payload) <= 0 {
		return types.ErrTxFormatInvalid
	}
	systemCmd := txBody.GetPayload()[0]
	var err error
	switch systemCmd {
	case 's':
		if txBody.Amount < types.StakingMinimum {
			return types.ErrTooSmallAmount
		}
	case 'v':
		/*
			TODO: need validate?
			peerID, err := peer.IDFromBytes(to)
			if err != nil {
				return err
			}
		*/
		_, when, _, err := getVote(scs, txBody.Account)
		if err != nil {
			return err
		}
		if when+VotingDelay > blockNo {
			//logger.Debug().Uint64("when", when).Uint64("blockNo", blockNo).Msg("remain voting delay")
			return types.ErrLessTimeHasPassed
		}
		staked, when, err := getStaking(scs, txBody.Account)
		if err != nil {
			return err
		}
		if staked == 0 {
			return types.ErrMustStakeBeforeVote
		}
		if when+VotingDelay > blockNo {
			//logger.Debug().Uint64("when", when).Uint64("blockNo", blockNo).Msg("remain voting delay")
			return types.ErrLessTimeHasPassed
		}
	case 'u':
		if txBody.Amount < types.StakingMinimum {
			return types.ErrTooSmallAmount
		}
		staked, when, err := getStaking(scs, txBody.Account)
		if err != nil {
			return err
		}
		if staked == 0 {
			return types.ErrMustStakeBeforeUnstake
		}
		if when+stakingDelay > blockNo {
			return types.ErrLessTimeHasPassed
		}
	}
	if err != nil {
		return err
	}
	return nil
}
