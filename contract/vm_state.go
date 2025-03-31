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

func getContractState(ctx *vmContext, id []byte) (*callState, error) {
	cs, err := getCallState(ctx, id)
	if err != nil {
		return nil, err
	}
	if cs.ctrState == nil {
		cs.ctrState, err = statedb.OpenContractState(id, cs.accState.State(), ctx.bs.StateDB)
	}
	return cs, err
}

func getOnlyContractState(ctx *vmContext, id []byte) (*statedb.ContractState, error) {
	cs := ctx.callState[types.ToAccountID(id)]
	if cs == nil || cs.ctrState == nil {
		return statedb.OpenContractStateAccount(id, ctx.bs.StateDB)
	}
	return cs.ctrState, nil
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

////////////////////////////////////////////////////////////////////////////////
// State Recovery
////////////////////////////////////////////////////////////////////////////////

// recoveryPoint is a struct that contains the state of the contract at a given point in time
type recoveryPoint struct {
	seq           int
	amount        *big.Int
	senderState   *state.AccountState
	senderNonce   uint64
	callState     *callState
	onlySend      bool
	isDeploy      bool
	sqlSaveName   *string
	stateRevision statedb.Snapshot
	prev          *recoveryPoint
}

// reverts the state of the contract to a previous state, stored on a recovery point
func (rp *recoveryPoint) revertState(bs *state.BlockState) error {
	var zero big.Int
	cs := rp.callState

	// restore the contract balance
	if rp.amount.Cmp(&zero) > 0 {
		if rp.senderState != nil {
			rp.senderState.AddBalance(rp.amount)
		}
		if cs != nil {
			cs.accState.SubBalance(rp.amount)
		}
	}
	if rp.onlySend {
		return nil
	}

	// restore the contract nonce
	if rp.senderState != nil {
		rp.senderState.SetNonce(rp.senderNonce)
	}

	// if the contract state is not stored, do not restore it
	if cs == nil {
		return nil
	}

	// restore the contract state
	if rp.stateRevision != -1 {
		err := cs.ctrState.Rollback(rp.stateRevision)
		if err != nil {
			return newDbSystemError(err)
		}
		if rp.isDeploy {
			err := cs.ctrState.SetCode(nil, nil)
			if err != nil {
				return newDbSystemError(err)
			}
			bs.RemoveCache(cs.ctrState.GetAccountID())
		}
	}

	// restore the contract SQL db state
	if cs.tx != nil {
		if rp.sqlSaveName == nil {
			err := cs.tx.rollbackToSavepoint()
			if err != nil {
				return newDbSystemError(err)
			}
			cs.tx = nil
		} else {
			err := cs.tx.rollbackToSubSavepoint(*rp.sqlSaveName)
			if err != nil {
				return newDbSystemError(err)
			}
		}
	}

	return nil
}

// creates a recovery point on the current state of the VM
func createRecoveryPoint(
	aid types.AccountID,
	ctx *vmContext,
	senderState *state.AccountState,
	cs *callState,
	amount *big.Int,
	onlySend, isDeploy bool,
) (int, error) {
	var seq int

	// get the previous recovery point
	prev := ctx.lastRecoveryPoint

	// get the next sequence number
	if prev != nil {
		seq = prev.seq + 1
	} else {
		seq = 1
	}

	// get the sender nonce
	var nonce uint64
	if senderState != nil {
		nonce = senderState.Nonce()
	}

	// create the recovery point
	rp := &recoveryPoint{
		seq,
		amount,
		senderState,
		nonce,
		cs,
		onlySend,
		isDeploy,
		nil,
		-1,
		prev,
	}

	// set this as the last recovery point
	ctx.lastRecoveryPoint = rp

	// if it's just aergo transfer, do not store the contract state
	if onlySend {
		return seq, nil
	}

	// get the contract state snapshot
	rp.stateRevision = cs.ctrState.Snapshot()

	// get the contract SQL db transaction
	tx := cs.tx
	if tx != nil {
		saveName := fmt.Sprintf("%s_%p", aid.String(), &rp)
		err := tx.subSavepoint(saveName)
		if err != nil {
			return seq, err
		}
		rp.sqlSaveName = &saveName
	}

	// return the sequence number
	return seq, nil
}

func getTraceFile(blkno uint64, tx []byte) *os.File {
	f, _ := os.OpenFile(fmt.Sprintf("%s%s%d.trace", os.TempDir(), string(os.PathSeparator), blkno), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		_, _ = f.WriteString(fmt.Sprintf("[START TX]: %s\n", base58.Encode(tx)))
	}
	return f
}
