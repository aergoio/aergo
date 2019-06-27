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

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

//SystemContext is context of executing aergo.system transaction and filled after validation.
type SystemContext struct {
	BlockNo  uint64
	Call     *types.CallInfo
	Args     []string
	Staked   *types.Staking
	Vote     *types.Vote // voting
	Proposal *Proposal   // voting
	Sender   *state.V
	Receiver *state.V

	op     types.OpSysTx
	scs    *state.ContractState
	txBody *types.TxBody
}

func newSystemContext(account []byte, txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockNo uint64) (*SystemContext, error) {
	context, err := ValidateSystemTx(sender.ID(), txBody, sender, scs, blockNo)
	if err != nil {
		return nil, err
	}
	context.Receiver = receiver

	return context, err
}

func (ctx *SystemContext) arg(i int) interface{} {
	return ctx.Call.Args[i]
}

type sysCmd interface {
	run() (*types.Event, error)
	arg(i int) interface{}
}

type sysCmdCtor func(ctx *SystemContext) (sysCmd, error)

func newSysCmd(account []byte, txBody *types.TxBody, sender, receiver *state.V,
	scs *state.ContractState, blockNo uint64) (sysCmd, error) {

	cmds := map[types.OpSysTx]sysCmdCtor{
		types.OpvoteBP:         newVoteCmd,
		types.OpvoteProposal:   newVoteCmd,
		types.Opstake:          newStakeCmd,
		types.Opunstake:        newUnstakeCmd,
		types.OpcreateProposal: newProposalCmd,
	}

	context, err := newSystemContext(account, txBody, sender, receiver, scs, blockNo)
	if err != nil {
		return nil, err
	}

	ctor, exist := cmds[types.GetOpSysTx(context.Call.Name)]
	if !exist {
		return nil, types.ErrTxInvalidPayload
	}

	return ctor(context)
}

func ExecuteSystemTx(scs *state.ContractState, txBody *types.TxBody,
	sender, receiver *state.V, blockNo types.BlockNo) ([]*types.Event, error) {

	cmd, err := newSysCmd(sender.ID(), txBody, sender, receiver, scs, blockNo)
	if err != nil {
		return nil, err
	}

	event, err := cmd.run()
	if err != nil {
		return nil, err
	}

	return []*types.Event{event}, nil
}

func GetNamePrice(scs *state.ContractState) *big.Int {
	//	votelist, err := getVoteResult(scs, []byte(types.VoteNamePrice[2:]), 1)
	//	if err != nil {
	//		panic("could not get vote result for min staking")
	//	}
	//	if len(votelist.Votes) == 0 {
	//		return types.NamePrice
	//	}
	//	return new(big.Int).SetBytes(votelist.Votes[0].GetCandidate())
	return types.NamePrice
}

func GetMinimumStaking(ar AccountStateReader) *big.Int {
	return types.StakingMinimum
}

func getMinimumStaking(scs *state.ContractState) *big.Int {
	//votelist, err := getVoteResult(scs, []byte(types.VoteMinStaking[2:]), 1)
	//if err != nil {
	//	panic("could not get vote result for min staking")
	//}
	//if len(votelist.Votes) == 0 {
	//	return types.StakingMinimum
	//}
	//minimumStaking, ok := new(big.Int).SetString(string(votelist.Votes[0].GetCandidate()), 10)
	//if !ok {
	//	panic("could not get vote result for min staking")
	//}
	//return minimumStaking
	return types.StakingMinimum
}

func GetVotes(scs *state.ContractState, address []byte) ([]*types.VoteInfo, error) {
	votes := getProposalHistory(scs, address)
	var results []*types.VoteInfo
	votes = append(votes, []byte(defaultVoteKey))
	for _, key := range votes {
		id := ProposalIDfromKey(key)
		result := &types.VoteInfo{Id: id}
		v, err := getVote(scs, key, address)
		if err != nil {
			return nil, err
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
