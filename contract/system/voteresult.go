package system

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/mr-tron/base58"
)

type VoteResult struct {
	rmap  map[string]*big.Int
	key   []byte
	ex    bool
	total *big.Int

	scs *state.ContractState
}

func newVoteResult(key []byte, total *big.Int) *VoteResult {
	voteResult := &VoteResult{}
	voteResult.rmap = map[string]*big.Int{}
	if bytes.Equal(key, defaultVoteKey) {
		voteResult.ex = false
	} else {
		voteResult.ex = true
		voteResult.total = total
	}
	voteResult.key = key
	return voteResult
}

func (voteResult *VoteResult) GetTotal() *big.Int {
	return voteResult.total
}

func (voteResult *VoteResult) SubVote(vote *types.Vote) error {
	if voteResult.ex {
		if vote.Candidate != nil {
			var args []string
			err := json.Unmarshal(vote.Candidate, &args)
			if err != nil {
				return err
			}
			for _, v := range args {
				voteResult.rmap[v] = new(big.Int).Sub(voteResult.rmap[v], vote.GetAmountBigInt())
			}
		}
		voteResult.total = new(big.Int).Sub(voteResult.total, vote.GetAmountBigInt())
	} else {
		for offset := 0; offset < len(vote.Candidate); offset += PeerIDLength {
			peer := vote.Candidate[offset : offset+PeerIDLength]
			pkey := base58.Encode(peer)
			voteResult.rmap[pkey] = new(big.Int).Sub(voteResult.rmap[pkey], vote.GetAmountBigInt())
		}
	}
	return nil
}

func (voteResult *VoteResult) AddVote(vote *types.Vote) error {
	if voteResult.ex {
		var args []string
		err := json.Unmarshal(vote.Candidate, &args)
		if err != nil {
			return err
		}
		for _, v := range args {
			if voteResult.rmap[v] == nil {
				voteResult.rmap[v] = new(big.Int).SetUint64(0)
			}
			voteResult.rmap[v] = new(big.Int).Add(voteResult.rmap[v], vote.GetAmountBigInt())
		}
		voteResult.total = new(big.Int).Add(voteResult.total, vote.GetAmountBigInt())
	} else {
		for offset := 0; offset < len(vote.Candidate); offset += PeerIDLength {
			key := vote.Candidate[offset : offset+PeerIDLength]
			if voteResult.rmap[base58.Encode(key)] == nil {
				voteResult.rmap[base58.Encode(key)] = new(big.Int).SetUint64(0)
			}
			voteResult.rmap[base58.Encode(key)] = new(big.Int).Add(voteResult.rmap[base58.Encode(key)], vote.GetAmountBigInt())
		}
	}
	return nil
}

func (vr *VoteResult) buildVoteList() *types.VoteList {
	var voteList types.VoteList
	for k, v := range vr.rmap {
		vote := &types.Vote{
			Amount: v.Bytes(),
		}
		if vr.ex {
			vote.Candidate = []byte(k)
		} else {
			vote.Candidate, _ = enc.ToBytes(k)
		}
		voteList.Votes = append(voteList.Votes, vote)
	}
	sort.Sort(sort.Reverse(voteList))

	return &voteList
}

// Sync is write vote result data to state DB. if vote result over the threshold,
func (vr *VoteResult) Sync() error {
	votingPowerRank.apply(vr.scs)
	resultList := vr.buildVoteList()
	if vr.ex {
		if vr.threshold(resultList.Votes[0].GetAmountBigInt()) {
			value, ok := new(big.Int).SetString(string(resultList.Votes[0].GetCandidate()), 10)
			if !ok {
				return fmt.Errorf("abnormal winner is in vote %s", string(vr.key))
			}
			if _, err := updateParam(vr.scs, string(vr.key), value); err != nil {
				return err
			}
		}
		if err := vr.scs.SetData(append(totalKey, vr.key...), vr.total.Bytes()); err != nil {
			return err
		}
	}
	return vr.scs.SetData(append(sortKey, vr.key...), serializeVoteList(resultList, vr.ex))
}

func (vr *VoteResult) threshold(power *big.Int) bool {
	if power.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	total, err := getStakingTotal(vr.scs)
	if err != nil {
		panic("failed to get staking total when calculate bp count")
	}
	if new(big.Int).Div(total, new(big.Int).Div(power, big.NewInt(100))).Cmp(big.NewInt(150)) <= 0 {
		return true
	}
	return false
}

func loadVoteResult(scs *state.ContractState, key []byte) (*VoteResult, error) {
	data, err := scs.GetData(append(sortKey, key...))
	if err != nil {
		return nil, err
	}
	total, err := scs.GetData(append(totalKey, key...))
	if err != nil {
		return nil, err
	}
	voteResult := newVoteResult(key, new(big.Int).SetBytes(total))
	if len(data) != 0 {
		voteList := deserializeVoteList(data, voteResult.ex)
		if voteList != nil {
			for _, v := range voteList.GetVotes() {
				if voteResult.ex {
					voteResult.rmap[string(v.Candidate)] = v.GetAmountBigInt()
				} else {
					voteResult.rmap[base58.Encode(v.Candidate)] = v.GetAmountBigInt()
				}
			}
		}
	}
	voteResult.scs = scs

	return voteResult, nil
}

func InitVoteResult(scs *state.ContractState, voteResult map[string]*big.Int) error {
	if voteResult == nil {
		return errors.New("Invalid argument : voteReult should not nil")
	}
	res := newVoteResult(defaultVoteKey, nil)
	res.rmap = voteResult
	res.scs = scs

	return res.Sync()
}

func getVoteResult(scs *state.ContractState, key []byte, n int) (*types.VoteList, error) {
	data, err := scs.GetData(append(sortKey, key...))
	if err != nil {
		return nil, err
	}
	var ex bool
	if bytes.Equal(key, defaultVoteKey) {
		ex = false
	} else {
		ex = true
	}
	voteList := deserializeVoteList(data, ex)
	if n < len(voteList.Votes) {
		voteList.Votes = voteList.Votes[:n]
	}
	return voteList, nil
}

func GetVoteResultEx(ar AccountStateReader, key []byte, n int) (*types.VoteList, error) {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		return nil, err
	}
	return getVoteResult(scs, key, n)
}
