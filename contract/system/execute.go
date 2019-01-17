/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"math"
	"math/big"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

const FutureBlockNo = math.MaxUint64
const lsep = byte(':')
const vsep = byte(',')

func ExecuteSystemTx(scs *state.ContractState, txBody *types.TxBody, sender *state.V,
	blockNo types.BlockNo) error {

	systemCmd, err := getSystemCmd(txBody.GetPayload())

	switch systemCmd {
	case 's':
		err = staking(txBody, sender, scs, blockNo)
	case 'v':
		err = voting(txBody, sender, scs, blockNo)
	case 'u':
		err = unstaking(txBody, sender, scs, blockNo)
	}
	if err != nil {
		return err
	}

	return nil
}

func ValidateSystemTx(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64) error {
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
		staked, err := getStaking(scs, account)
		if err != nil {
			return err
		}
		if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
			return types.ErrMustStakeBeforeVote
		}
		oldvote, err := getVote(scs, account)
		if err != nil {
			return err
		}
		if oldvote.Amount != nil && staked.GetWhen()+VotingDelay > blockNo {
			//logger.Debug().Uint64("when", when).Uint64("blockNo", blockNo).Msg("remain voting delay")
			return types.ErrLessTimeHasPassed
		}
	case 'u':
		_, err = validateForUnstaking(account, txBody, scs, blockNo)
	}
	if err != nil {
		return err
	}
	return nil
}

func validateForStaking(txBody *types.TxBody, scs *state.ContractState, blockNo uint64) error {
	amount := txBody.GetAmountBigInt()
	if amount.Cmp(types.StakingMinimum) < 0 {
		return types.ErrTooSmallAmount
	}
	return nil
}

func validateForUnstaking(account []byte, txBody *types.TxBody, scs *state.ContractState, blockNo uint64) (*types.Staking, error) {
	amount := txBody.GetAmountBigInt()
	if amount.Cmp(types.StakingMinimum) < 0 {
		return nil, types.ErrTooSmallAmount
	}
	staked, err := getStaking(scs, account)
	if err != nil {
		return nil, err
	}
	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil, types.ErrMustStakeBeforeUnstake
	}
	if staked.GetWhen()+StakingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}
	return staked, nil
}

func getSystemCmd(payload []byte) (byte, error) {
	if len(payload) <= 0 {
		return 0, types.ErrTxFormatInvalid
	}
	return payload[0], nil
}
