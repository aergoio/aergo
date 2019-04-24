package system

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"sort"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

type VoteResult struct {
	rmap  map[string]*big.Int
	key   []byte
	ex    bool
	total *big.Int
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

func (vr *VoteResult) Sync(scs *state.ContractState) error {
	if vr.ex {
		if err := scs.SetData(append(totalKey, vr.key...), vr.total.Bytes()); err != nil {
			return err
		}
	}
	return scs.SetData(append(sortKey, vr.key...), serializeVoteList(vr.buildVoteList(), vr.ex))
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
	return voteResult, nil
}

func InitVoteResult(scs *state.ContractState, voteResult map[string]*big.Int) error {
	if voteResult == nil {
		return errors.New("Invalid argument : voteReult should not nil")
	}
	res := newVoteResult(defaultVoteKey, nil)
	res.rmap = voteResult
	return res.Sync(scs)
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
