/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"math"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

const FutureBlockNo = math.MaxUint64

func ExecuteSystemTx(txBody *types.TxBody, senderState *types.State,
	scs *state.ContractState, blockNo types.BlockNo) error {

	systemCmd, err := getSystemCmd(txBody.GetPayload())

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

func ValidateSystemTx(txBody *types.TxBody, scs *state.ContractState, blockNo uint64) error {
	systemCmd, err := getSystemCmd(txBody.GetPayload())
	switch systemCmd {
	case 's':
		err = validateForStaking(txBody, scs, blockNo)
	case 'v':
		if len(txBody.Payload[1:])%PeerIDLength != 0 {
			return types.ErrTxFormatInvalid
		}
		for offset := 0; offset < len(txBody.Payload[1:]); offset += PeerIDLength {
			_, err := peer.IDFromBytes(txBody.Payload[offset+1 : offset+PeerIDLength+1])
			if err != nil {
				return err
			}
		}
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
		_, _, err = validateForUnstaking(txBody, scs, blockNo)
	}
	if err != nil {
		return err
	}
	return nil
}

func validateForStaking(txBody *types.TxBody, scs *state.ContractState, blockNo uint64) error {
	if txBody.Amount < types.StakingMinimum {
		return types.ErrTooSmallAmount
	}
	return nil
}

func validateForUnstaking(txBody *types.TxBody, scs *state.ContractState, blockNo uint64) (uint64, uint64, error) {
	if txBody.Amount < types.StakingMinimum {
		return 0, 0, types.ErrTooSmallAmount
	}
	staked, when, err := getStaking(scs, txBody.Account)
	if err != nil {
		return 0, 0, err
	}
	if staked == 0 {
		return 0, 0, types.ErrMustStakeBeforeUnstake
	}
	if when+StakingDelay > blockNo {
		return 0, 0, types.ErrLessTimeHasPassed
	}
	return staked, when, nil
}

func getSystemCmd(payload []byte) (byte, error) {
	if len(payload) <= 0 {
		return 0, types.ErrTxFormatInvalid
	}
	return payload[0], nil
}
