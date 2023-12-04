package contract

import (
	"fmt"
	"math/big"
	"os"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
)

type callState struct {
	ctrState  *state.ContractState
	prevState *types.State
	curState  *types.State
	tx        sqlTx
}

func getCallState(ctx *vmContext, id []byte) (*callState, error) {
	aid := types.ToAccountID(id)
	cs := ctx.callState[aid]
	if cs == nil {
		bs := ctx.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			return nil, err
		}

		curState := prevState.Clone()
		cs = &callState{prevState: prevState, curState: curState}
		ctx.callState[aid] = cs
	}
	return cs, nil
}

func getCtrState(ctx *vmContext, id []byte) (*callState, error) {
	cs, err := getCallState(ctx, id)
	if err != nil {
		return nil, err
	}
	if cs.ctrState == nil {
		cs.ctrState, err = state.OpenContractState(id, cs.curState, ctx.bs.StateDB)
	}
	return cs, err
}

type contractInfo struct {
	callState  *callState
	sender     []byte
	contractId []byte
	rp         uint64
	amount     *big.Int
}

func newContractInfo(cs *callState, sender, contractId []byte, rp uint64, amount *big.Int) *contractInfo {
	return &contractInfo{
		cs,
		sender,
		contractId,
		rp,
		amount,
	}
}

func getOnlyContractState(ctx *vmContext, id []byte) (*state.ContractState, error) {
	cs := ctx.callState[types.ToAccountID(id)]
	if cs == nil || cs.ctrState == nil {
		return state.OpenContractStateAccount(id, ctx.bs.StateDB)
	}
	return cs.ctrState, nil
}

type recoveryEntry struct {
	seq           int
	amount        *big.Int
	senderState   *types.State
	senderNonce   uint64
	callState     *callState
	onlySend      bool
	isDeploy      bool
	sqlSaveName   *string
	stateRevision statedb.Snapshot
	prev          *recoveryEntry
}

func (re *recoveryEntry) recovery(bs *state.BlockState) error {
	var zero big.Int
	cs := re.callState
	if re.amount.Cmp(&zero) > 0 {
		if re.senderState != nil {
			re.senderState.Balance = new(big.Int).Add(re.senderState.GetBalanceBigInt(), re.amount).Bytes()
		}
		if cs != nil {
			cs.curState.Balance = new(big.Int).Sub(cs.curState.GetBalanceBigInt(), re.amount).Bytes()
		}
	}
	if re.onlySend {
		return nil
	}
	if re.senderState != nil {
		re.senderState.Nonce = re.senderNonce
	}

	if cs == nil {
		return nil
	}
	if re.stateRevision != -1 {
		err := cs.ctrState.Rollback(re.stateRevision)
		if err != nil {
			return newDbSystemError(err)
		}
		if re.isDeploy {
			err := cs.ctrState.SetCode(nil)
			if err != nil {
				return newDbSystemError(err)
			}
			bs.RemoveCache(cs.ctrState.GetAccountID())
		}
	}
	if cs.tx != nil {
		if re.sqlSaveName == nil {
			err := cs.tx.rollbackToSavepoint()
			if err != nil {
				return newDbSystemError(err)
			}
			cs.tx = nil
		} else {
			err := cs.tx.rollbackToSubSavepoint(*re.sqlSaveName)
			if err != nil {
				return newDbSystemError(err)
			}
		}
	}
	return nil
}

func getTraceFile(blkno uint64, tx []byte) *os.File {
	f, _ := os.OpenFile(fmt.Sprintf("%s%s%d.trace", os.TempDir(), string(os.PathSeparator), blkno), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		_, _ = f.WriteString(fmt.Sprintf("[START TX]: %s\n", base58.Encode(tx)))
	}
	return f
}
