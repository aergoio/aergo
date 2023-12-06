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
	isCallback bool
	isDeploy   bool

	ctrState *statedb.ContractState
	accState *state.AccountState
	tx       sqlTx
}

func getCallState(ctx *vmContext, id []byte) (*callState, error) {
	aid := types.ToAccountID(id)
	cs := ctx.callState[aid]
	if cs == nil {
		bs := ctx.bs
		accState, err := state.GetAccountState(id, bs.StateDB)
		if err != nil {
			return nil, err
		}
		cs = &callState{isCallback: true, accState: accState}
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
		cs.ctrState, err = statedb.OpenContractState(id, cs.accState.State(), ctx.bs.StateDB)
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

func getOnlyContractState(ctx *vmContext, id []byte) (*statedb.ContractState, error) {
	cs := ctx.callState[types.ToAccountID(id)]
	if cs == nil || cs.ctrState == nil {
		return statedb.OpenContractStateAccount(id, ctx.bs.StateDB)
	}
	return cs.ctrState, nil
}

type recoveryEntry struct {
	seq           int
	amount        *big.Int
	senderState   *state.AccountState
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
			re.senderState.AddBalance(re.amount)
		}
		if cs != nil {
			cs.accState.SubBalance(re.amount)
		}
	}
	if re.onlySend {
		return nil
	}
	if re.senderState != nil {
		re.senderState.SetNonce(re.senderNonce)
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

func setRecoveryPoint(aid types.AccountID, ctx *vmContext, senderState *state.AccountState,
	cs *callState, amount *big.Int, isSend, isDeploy bool) (int, error) {
	var seq int
	prev := ctx.lastRecoveryEntry
	if prev != nil {
		seq = prev.seq + 1
	} else {
		seq = 1
	}
	var nonce uint64
	if senderState != nil {
		nonce = senderState.Nonce()
	}
	re := &recoveryEntry{
		seq,
		amount,
		senderState,
		nonce,
		cs,
		isSend,
		isDeploy,
		nil,
		-1,
		prev,
	}
	ctx.lastRecoveryEntry = re
	if isSend {
		return seq, nil
	}
	re.stateRevision = cs.ctrState.Snapshot()
	tx := cs.tx
	if tx != nil {
		saveName := fmt.Sprintf("%s_%p", aid.String(), &re)
		err := tx.subSavepoint(saveName)
		if err != nil {
			return seq, err
		}
		re.sqlSaveName = &saveName
	}
	return seq, nil
}

func getTraceFile(blkno uint64, tx []byte) *os.File {
	f, _ := os.OpenFile(fmt.Sprintf("%s%s%d.trace", os.TempDir(), string(os.PathSeparator), blkno), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		_, _ = f.WriteString(fmt.Sprintf("[START TX]: %s\n", base58.Encode(tx)))
	}
	return f
}
