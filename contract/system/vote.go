/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

var lastBpCount int

var voteKey = []byte("vote")
var totalKey = []byte("total")
var sortKey = []byte("sort")

const PeerIDLength = 39

const VotingDelay = 60 * 60 * 24 //block interval
//const VotingDelay = 5

var defaultVoteKey = []byte(types.VoteBP)[2:]

func voting(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	var key []byte
	var args []byte
	var err error
	if context.Proposal != nil {
		key = context.Proposal.GetKey()
		args, err = json.Marshal(context.Call.Args[1:]) //[0] is name
		if err != nil {
			return nil, err
		}
		if err := addProposalHistory(scs, sender.ID(), context.Proposal); err != nil {
			return nil, err
		}
	} else {
		key = []byte(context.Call.Name)[2:]
		args, err = json.Marshal(context.Call.Args)
		if err != nil {
			return nil, err
		}
	}
	oldvote := context.Vote
	staked := context.Staked
	//update block number
	staked.When = blockNo
	err = setStaking(scs, sender.ID(), staked)
	if err != nil {
		return nil, err
	}

	voteResult, err := loadVoteResult(scs, key)
	if err != nil {
		return nil, err
	}

	err = voteResult.SubVote(oldvote)
	if err != nil {
		return nil, err
	}

	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil, types.ErrMustStakeBeforeVote
	}
	vote := &types.Vote{Amount: staked.GetAmount()}
	var candidates []byte
	if bytes.Equal(key, defaultVoteKey) {
		for _, v := range context.Call.Args {
			candidate, _ := base58.Decode(v.(string))
			candidates = append(candidates, candidate...)
		}
		vote.Candidate = candidates
	} else {
		vote.Candidate = args
	}

	err = setVote(scs, key, sender.ID(), vote)
	if err != nil {
		return nil, err
	}
	err = voteResult.AddVote(vote)
	if err != nil {
		return nil, err
	}

	err = voteResult.Sync(scs)
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       context.Call.Name[2:],
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "vote":` + string(args) + `}`,
	}, nil
}

func refreshAllVote(txBody *types.TxBody, scs *state.ContractState,
	context *SystemContext) error {
	account := context.Sender.ID()
	staked := context.Staked
	stakedAmount := new(big.Int).SetBytes(staked.Amount)
	allVotes := getProposalHistory(scs, account)
	allVotes = append(allVotes, []byte(types.VoteBP[2:]))
	for _, key := range allVotes {
		oldvote, err := getVote(scs, key, account)
		if err != nil {
			return err
		}
		if oldvote.Amount == nil ||
			new(big.Int).SetBytes(oldvote.Amount).Cmp(stakedAmount) <= 0 {
			continue
		}
		proposal, err := getProposal(scs, types.ProposalIDfromKey(key))
		if err != nil {
			return err
		}
		if proposal != nil && proposal.GetBlockto() < context.BlockNo {
			continue
		}
		voteResult, err := loadVoteResult(scs, key)
		if err != nil {
			return err
		}
		if err = voteResult.SubVote(oldvote); err != nil {
			return err
		}
		oldvote.Amount = staked.GetAmount()
		if err = setVote(scs, key, account, oldvote); err != nil {
			return err
		}
		if err = voteResult.AddVote(oldvote); err != nil {
			return err
		}
		if err = voteResult.Sync(scs); err != nil {
			return err
		}
	}
	return nil
}

//GetVote return amount, to, err
func GetVote(scs *state.ContractState, voter []byte, title []byte) (*types.Vote, error) {
	return getVote(scs, title, voter)
}

func getVote(scs *state.ContractState, key, voter []byte) (*types.Vote, error) {
	dataKey := append(append(voteKey, key...), voter...)
	data, err := scs.GetData(dataKey)
	if err != nil {
		return nil, err
	}
	var vote types.Vote
	if len(data) != 0 {
		if bytes.Equal(key, defaultVoteKey) {
			return deserializeVote(data), nil
		} else {
			return deserializeVoteEx(data), nil
		}
	}

	return &vote, nil
}

func setVote(scs *state.ContractState, key, voter []byte, vote *types.Vote) error {
	dataKey := append(append(voteKey, key...), voter...)
	if bytes.Equal(key, defaultVoteKey) {
		return scs.SetData(dataKey, serializeVote(vote))
	} else {
		return scs.SetData(dataKey, serializeVoteEx(vote))
	}
}

// BuildOrderedCandidates returns a candidate list ordered by votes.xs
func BuildOrderedCandidates(vote map[string]*big.Int) []string {
	// TODO: cleanup
	voteResult := newVoteResult(defaultVoteKey, nil)
	voteResult.rmap = vote
	l := voteResult.buildVoteList()
	bps := make([]string, 0, len(l.Votes))
	for _, v := range l.Votes {
		bp := enc.ToString(v.Candidate)
		bps = append(bps, bp)
	}
	return bps
}

// AccountStateReader is an interface for getting a system account state.
type AccountStateReader interface {
	GetSystemAccountState() (*state.ContractState, error)
}

// GetVoteResult returns the top n voting result from the system account state.
func GetVoteResult(ar AccountStateReader, id []byte, n int) (*types.VoteList, error) {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	return getVoteResult(scs, id, n)
}

// InitDefaultBpCount sets lastBpCount to bpCount.
//
// Caution: This function must be called only once before all the aergosvr
// services start.
func InitDefaultBpCount(bpCount int) {
	// Ensure that it is not modified after it is initialized.
	if lastBpCount > 0 {
		return
	}
	lastBpCount = bpCount
}

func getLastBpCount() int {
	return lastBpCount
}

func GetBpCount(ar AccountStateReader) int {
	result, err := GetVoteResultEx(ar, types.GenProposalKey("BPCOUNT"), 1)
	if err != nil {
		panic("could not get vote result for min staking")
	}
	if len(result.Votes) == 0 {
		return getLastBpCount()
	}
	power := result.Votes[0].GetAmountBigInt()
	if power.Cmp(big.NewInt(0)) == 0 {
		return getLastBpCount()
	}
	total, err := getTotal(ar)
	if err != nil {
		panic("failed to get staking total when calculate bp count")
	}
	if new(big.Int).Div(total, new(big.Int).Div(power, big.NewInt(100))).Cmp(big.NewInt(150)) <= 0 {
		bpcount, err := strconv.Atoi(string(result.Votes[0].GetCandidate()))
		if err != nil {
			return getLastBpCount()
		}
		lastBpCount = bpcount
		return bpcount
	}
	return getLastBpCount()
}

// GetRankers returns the IDs of the top n rankers.
func GetRankers(ar AccountStateReader) ([]string, error) {
	n := GetBpCount(ar)

	vl, err := GetVoteResult(ar, defaultVoteKey, n)
	if err != nil {
		return nil, err
	}

	bps := make([]string, 0, n)
	for _, v := range vl.Votes {
		bps = append(bps, enc.ToString(v.Candidate))
	}
	return bps, nil
}

func serializeVoteList(vl *types.VoteList, ex bool) []byte {
	var data []byte
	for _, v := range vl.GetVotes() {
		var serialized []byte
		if ex {
			serialized = serializeVoteEx(v)
		} else {
			serialized = serializeVote(v)
		}
		vsize := make([]byte, 8)
		binary.LittleEndian.PutUint64(vsize, uint64(len(serialized)))
		data = append(data, vsize...)
		data = append(data, serialized...)
	}
	return data
}

func serializeVote(v *types.Vote) []byte {
	var ret []byte
	if v != nil {
		ret = append(ret, v.GetCandidate()...)
		ret = append(ret, v.GetAmount()...)
	}
	return ret
}

func serializeVoteEx(v *types.Vote) []byte {
	var ret []byte
	if v != nil {
		size := make([]byte, 8)
		binary.LittleEndian.PutUint64(size, uint64(len(v.Candidate)))
		ret = append(ret, size...)
		ret = append(ret, v.GetCandidate()...)
		ret = append(ret, v.GetAmount()...)
	}
	return ret
}

func deserializeVote(data []byte) *types.Vote {
	pos := len(data) % PeerIDLength
	candidate := data[:len(data)-pos]
	amount := data[len(data)-pos:]
	if len(candidate)%PeerIDLength != 0 {
		panic("voting data corruption")
	}
	return &types.Vote{Amount: amount, Candidate: candidate}
}

func deserializeVoteEx(data []byte) *types.Vote {
	size := int(binary.LittleEndian.Uint64(data[:8]))
	candidate := data[8 : 8+size]
	amount := data[8+size:]
	return &types.Vote{Amount: amount, Candidate: candidate}
}

func deserializeVoteList(data []byte, ex bool) *types.VoteList {
	vl := &types.VoteList{Votes: []*types.Vote{}}
	var end int
	for offset := 0; offset < len(data); offset = end {
		size := binary.LittleEndian.Uint64(data[offset : offset+8])
		end = offset + 8 + int(size)
		v := data[offset+8 : end]
		if ex {
			vl.Votes = append(vl.Votes, deserializeVoteEx(v))
		} else {
			vl.Votes = append(vl.Votes, deserializeVote(v))
		}
	}
	return vl
}
