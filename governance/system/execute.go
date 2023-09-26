/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58"
)

// SystemContext is context of executing aergo.system transaction and filled after validation.
type SystemContext struct {
	BlockInfo *types.BlockHeaderInfo
	Call      *types.CallInfo
	Args      []string
	Staked    *types.Staking
	Vote      *types.Vote // voting
	Proposal  *Proposal   // voting
	Sender    *state.V
	Receiver  *state.V

	op     types.OpSysTx
	scs    *state.ContractState
	txBody *types.TxBody
}

func NewSystemContext(proposals *Proposals, account []byte, txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockInfo *types.BlockHeaderInfo) (*SystemContext, error) {
	context, err := ValidateSystemTx(proposals, sender.ID(), txBody, sender, scs, blockInfo)
	if err != nil {
		return nil, err
	}
	context.Receiver = receiver

	return context, err
}

func (ctx *SystemContext) arg(i int) interface{} {
	return ctx.Call.Args[i]
}

// Update the sender's staking.
func (c *SystemContext) updateStaking() error {
	return setStaking(c.scs, c.Sender.ID(), c.Staked)
}

type SysCmd interface {
	run() (*types.Event, error)
	arg(i int) interface{}
}

type SysCmdCtor func(ctx *SystemContext) (SysCmd, error)

func newSysCmd(proposals *Proposals, account []byte, txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockInfo *types.BlockHeaderInfo) (SysCmd, error) {

	cmds := map[types.OpSysTx]SysCmdCtor{
		types.OpvoteBP:  NewVoteCmd,
		types.OpvoteDAO: NewVoteCmd,
		types.Opstake:   NewStakeCmd,
		types.Opunstake: NewUnstakeCmd,
	}

	context, err := NewSystemContext(proposals, account, txBody, sender, receiver, scs, blockInfo)
	if err != nil {
		return nil, err
	}

	ctor, exist := cmds[types.GetOpSysTx(context.Call.Name)]
	if !exist {
		return nil, types.ErrTxInvalidPayload
	}

	return ctor(context)
}

func ExecuteSystemTx(proposals *Proposals, scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockInfo *types.BlockHeaderInfo) ([]*types.Event, error) {

	cmd, err := newSysCmd(proposals, sender.ID(), txBody, sender, receiver, scs, blockInfo)
	if err != nil {
		return nil, err
	}

	event, err := cmd.run()
	if err != nil {
		return nil, err
	}

	return []*types.Event{event}, nil
}

func GetVotes(scs *state.ContractState, votingCatalog []types.VotingIssue, address []byte) ([]*types.VoteInfo, error) {
	var results []*types.VoteInfo

	for _, i := range votingCatalog {
		id := i.ID()
		key := i.Key()

		result := &types.VoteInfo{Id: id}
		v, err := getVote(scs, key, address)
		if err != nil {
			return nil, err
		}
		if v.Amount == nil {
			continue
		}

		if bytes.Equal(key, defaultVoteKey) {
			for offset := 0; offset < len(v.Candidate); offset += PeerIDLength {
				candi := base58.Encode(v.Candidate[offset : offset+PeerIDLength])
				result.Candidates = append(result.Candidates, candi)
			}
		} else {
			err := json.Unmarshal(v.Candidate, &result.Candidates)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", err.Error(), string(v.Candidate))
			}
		}
		result.Amount = new(big.Int).SetBytes(v.Amount).String()
		results = append(results, result)
	}
	return results, nil
}
