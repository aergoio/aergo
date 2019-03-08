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
	"sort"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

var voteKey = []byte("vote")
var sortKey = []byte("sort")

const PeerIDLength = 39
const VotingDelay = 60 * 60 * 24 //block interval

var defaultVoteKey = []byte(types.VoteBP)[2:]

func voting(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, ci *types.CallInfo) (*types.Event, error) {

	fromUnstake := false

	var key []byte
	if ci.Name == types.Unstake { //called from unstaking
		fromUnstake = true
		key = ci.Args[0].([]byte)
	} else {
		key = []byte(ci.Name)[2:]
	}

	oldvote, err := getVote(scs, key, sender.ID())
	if err != nil {
		return nil, err
	}

	staked, err := getStaking(scs, sender.ID())
	if err != nil {
		return nil, err
	}

	if oldvote.Amount != nil && staked.GetWhen()+VotingDelay > blockNo {
		return nil, types.ErrLessTimeHasPassed
	}

	staked.When = blockNo
	err = setStaking(scs, sender.ID(), staked)
	if err != nil {
		return nil, err
	}

	voteResult, err := loadVoteResult(scs, key)
	if err != nil {
		return nil, err
	}

	voteResult.SubVote(oldvote)
	var args []byte
	if fromUnstake { //called from unstaking
		oldvote.Amount = staked.GetAmount()
		err = setVote(scs, key, sender.ID(), oldvote)
		if err != nil {
			return nil, err
		}
		voteResult.AddVote(oldvote)
	} else {
		args, err = json.Marshal(ci.Args)
		if err != nil {
			return nil, err
		}
		if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
			return nil, types.ErrMustStakeBeforeVote
		}
		vote := &types.Vote{Amount: staked.GetAmount()}
		var candidates []byte
		if bytes.Equal(key, defaultVoteKey) {
			for _, v := range ci.Args {
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
		voteResult.AddVote(vote)
	}

	err = voteResult.Sync(scs)
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       ci.Name[2:],
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "vote":` + string(args) + `}`,
	}, nil
}

//GetVote return amount, to, err
func GetVote(scs *state.ContractState, voter []byte) (*types.Vote, error) {
	return getVote(scs, defaultVoteKey, voter)
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
	l := buildVoteList(vote)
	bps := make([]string, 0, len(l.Votes))
	for _, v := range l.Votes {
		bp := enc.ToString(v.Candidate)
		bps = append(bps, bp)
	}
	return bps
}

// BuildVoteList builds and returns a voteList type obejct from vote.
func buildVoteList(vote map[string]*big.Int) *types.VoteList {
	var voteList types.VoteList
	for k, v := range vote {
		c, _ := enc.ToBytes(k)
		vote := &types.Vote{
			Candidate: c,
			Amount:    v.Bytes(),
		}
		voteList.Votes = append(voteList.Votes, vote)
	}
	sort.Sort(sort.Reverse(voteList))

	return &voteList
}

// AccountStateReader is an interface for getting a system account state.
type AccountStateReader interface {
	GetSystemAccountState() (*state.ContractState, error)
}

// GetVoteResult returns the top n voting result from the system account state.
func GetVoteResult(ar AccountStateReader, n int) (*types.VoteList, error) {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	return getVoteResult(scs, defaultVoteKey, n)
}

// GetRankers returns the IDs of the top n rankers.
func GetRankers(ar AccountStateReader, n int) ([]string, error) {
	vl, err := GetVoteResult(ar, n)
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
	size := binary.LittleEndian.Uint64(data[:8])
	pos := len(data) % int(size)
	candidate := data[:len(data)-pos]
	amount := data[len(data)-pos:]
	if len(candidate)%int(size) != 0 {
		panic("voting data ex corruption")
	}
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
