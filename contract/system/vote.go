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
	key := []byte(ci.Name)[2:]
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
	args, err := json.Marshal(ci.Args)
	if err != nil {
		return nil, err
	}
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
		EventName:       ci.Name[2:],
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "vote":` + string(args) + `}`,
	}, nil
}

func refreshAllVote(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo) error {
	for _, keystr := range types.AllVotes {
		key := []byte(keystr[2:])
		oldvote, err := getVote(scs, key, sender.ID())
		if err != nil {
			return err
		}
		if oldvote.Amount == nil {
			continue
		}
		voteResult, err := loadVoteResult(scs, key)
		if err != nil {
			return err
		}
		if err = voteResult.SubVote(oldvote); err != nil {
			return err
		}

		staked, err := getStaking(scs, sender.ID())
		if err != nil {
			return err
		}
		oldvote.Amount = staked.GetAmount()
		if err = setVote(scs, key, sender.ID(), oldvote); err != nil {
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
	voteResult := newVoteResult(defaultVoteKey)
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

// GetRankers returns the IDs of the top n rankers.
func GetRankers(ar AccountStateReader, n int) ([]string, error) {
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
