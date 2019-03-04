/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"encoding/binary"
	"encoding/json"
	"errors"
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

	for offset := 0; offset < len(oldvote.Candidate); offset += PeerIDLength {
		key := oldvote.Candidate[offset : offset+PeerIDLength]
		voteResult[base58.Encode(key)] = new(big.Int).Sub(voteResult[base58.Encode(key)], oldvote.GetAmountBigInt())
	}

	if fromUnstake { //called from unstaking
		oldvote.Amount = staked.GetAmount()
		err = setVote(scs, key, sender.ID(), oldvote)
		if err != nil {
			return nil, err
		}
		for offset := 0; offset < len(oldvote.Candidate); offset += PeerIDLength {
			key := oldvote.Candidate[offset : offset+PeerIDLength]
			voteResult[base58.Encode(key)] = new(big.Int).Add(voteResult[base58.Encode(key)], staked.GetAmountBigInt())
		}
	} else {
		if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
			return nil, types.ErrMustStakeBeforeVote
		}
		var candidates []byte
		for _, v := range ci.Args {
			candidate, _ := base58.Decode(v.(string))
			candidates = append(candidates, candidate...)
		}
		vote := &types.Vote{Candidate: candidates, Amount: staked.GetAmount()}
		err = setVote(scs, key, sender.ID(), vote)
		if err != nil {
			return nil, err
		}
		for offset := 0; offset < len(candidates); offset += PeerIDLength {
			key := candidates[offset : offset+PeerIDLength]

			if voteResult[base58.Encode(key)] == nil {
				voteResult[base58.Encode(key)] = new(big.Int).SetUint64(0)
			}

			voteResult[base58.Encode(key)] = new(big.Int).Add(voteResult[base58.Encode(key)], staked.GetAmountBigInt())
		}
	}

	err = syncVoteResult(scs, key, voteResult)
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(ci.Args)
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       ci.Name[2:],
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "vote":` + string(result) + `}`,
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
		return deserializeVote(data), nil
	}

	return &vote, nil
}

func setVote(scs *state.ContractState, key, voter []byte, vote *types.Vote) error {
	dataKey := append(append(voteKey, key...), voter...)
	return scs.SetData(dataKey, serializeVote(vote))
}

func loadVoteResult(scs *state.ContractState, key []byte) (map[string]*big.Int, error) {
	voteResult := map[string]*big.Int{}
	data, err := scs.GetData(append(sortKey, key...))
	if err != nil {
		return nil, err
	}
	if len(data) != 0 {
		voteList := deserializeVoteList(data)
		if voteList != nil {
			for _, v := range voteList.GetVotes() {
				voteResult[base58.Encode(v.Candidate)] = v.GetAmountBigInt()
			}
		}
	}
	return voteResult, nil
}

func InitVoteResult(scs *state.ContractState, voteResult map[string]*big.Int) error {
	if voteResult == nil {
		return errors.New("Invalid argument : voteReult should not nil")
	}
	return syncVoteResult(scs, defaultVoteKey, voteResult)
}

func syncVoteResult(scs *state.ContractState, key []byte, voteResult map[string]*big.Int) error {
	voteList := buildVoteList(voteResult)

	//logger.Info().Msgf("VOTE set list %v", voteList.Votes)
	return scs.SetData(append(sortKey, key...), serializeVoteList(voteList))
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

func getVoteResult(scs *state.ContractState, key []byte, n int) (*types.VoteList, error) {
	data, err := scs.GetData(append(sortKey, key...))
	if err != nil {
		return nil, err
	}
	voteList := deserializeVoteList(data)
	if n < len(voteList.Votes) {
		voteList.Votes = voteList.Votes[:n]
	}
	return voteList, nil
}

func serializeVoteList(vl *types.VoteList) []byte {
	var data []byte
	for _, v := range vl.GetVotes() {
		v := serializeVote(v)
		vsize := make([]byte, 8)
		binary.LittleEndian.PutUint64(vsize, uint64(len(v)))
		data = append(data, vsize...)
		data = append(data, v...)
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

func deserializeVote(data []byte) *types.Vote {
	pos := len(data) % PeerIDLength
	candidate := data[:len(data)-pos]
	amount := data[len(data)-pos:]
	if len(candidate)%PeerIDLength != 0 {
		panic("voting data corruption")
	}
	return &types.Vote{Amount: amount, Candidate: candidate}
}

func deserializeVoteList(data []byte) *types.VoteList {
	vl := &types.VoteList{Votes: []*types.Vote{}}
	var end int
	for offset := 0; offset < len(data); offset = end {
		size := binary.LittleEndian.Uint64(data[offset : offset+8])
		end = offset + 8 + int(size)
		v := data[offset+8 : end]
		vl.Votes = append(vl.Votes, deserializeVote(v))
	}
	return vl
}
