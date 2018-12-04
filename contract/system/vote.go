/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package system

import (
	"bytes"
	"encoding/gob"
	"errors"
	"sort"

	"github.com/aergoio/aergo/internal/common"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58"
)

var votingkey = []byte("voting")
var totalkey = []byte("totalvote")
var sortedlistkey = []byte("sortedlist")

const PeerIDLength = 39
const VotingDelay = 5

func voting(txBody *types.TxBody, scs *state.ContractState, blockNo types.BlockNo) error {
	staked, err := getStaking(scs, txBody.Account)
	if err != nil {
		return err
	}
	if staked.GetWhen()+VotingDelay > blockNo {
		return types.ErrLessTimeHasPassed
	}
	staked.When = blockNo
	err = setStaking(scs, txBody.Account, staked)
	if err != nil {
		return err
	}
	oldvote, err := getVote(scs, txBody.Account)
	if err != nil {
		return err
	}

	voteResult, err := loadVoteResult(scs)
	if err != nil {
		return err
	}

	for offset := 0; offset < len(oldvote.Candidate); offset += PeerIDLength {
		key := oldvote.Candidate[offset : offset+PeerIDLength]
		(*voteResult)[base58.Encode(key)] -= oldvote.Amount
	}

	if txBody.Payload[0] != 'v' { //called from unstaking
		oldvote.Amount = staked.GetAmount()
		err = setVote(scs, txBody.Account, oldvote)
		if err != nil {
			return err
		}
		for offset := 0; offset < len(oldvote.Candidate); offset += PeerIDLength {
			key := oldvote.Candidate[offset : offset+PeerIDLength]
			(*voteResult)[base58.Encode(key)] += staked.GetAmount()
		}
	} else {
		if staked.GetAmount() == 0 {
			return types.ErrMustStakeBeforeVote
		}
		vote := &types.Vote{Candidate: txBody.Payload[1:], Amount: staked.GetAmount()}
		err = setVote(scs, txBody.Account, vote)
		if err != nil {
			return err
		}
		for offset := 0; offset < len(txBody.Payload[1:]); offset += PeerIDLength {
			key := txBody.Payload[offset+1 : offset+PeerIDLength+1]
			(*voteResult)[base58.Encode(key)] += staked.GetAmount()
		}
	}

	err = syncVoteResult(scs, voteResult)
	if err != nil {
		return err
	}
	return nil
}

//GetVote return amount, to, err
func GetVote(scs *state.ContractState, voter []byte) (*types.Vote, error) {
	return getVote(scs, voter)
}

func getVote(scs *state.ContractState, voter []byte) (*types.Vote, error) {
	key := append(votingkey, voter...)
	data, err := scs.GetData(key)
	if err != nil {
		return nil, err
	}
	var vote types.Vote
	if len(data) != 0 {
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		err = dec.Decode(&vote)
		if err != nil {
			return nil, err
		}
	}

	return &vote, nil
}

func setVote(scs *state.ContractState, voter []byte, vote *types.Vote) error {
	var data bytes.Buffer
	key := append(votingkey, voter...)
	enc := gob.NewEncoder(&data)
	err := enc.Encode(vote)
	if err != nil {
		return err
	}
	return scs.SetData(key, data.Bytes())
}

func loadVoteResult(scs *state.ContractState) (*map[string]uint64, error) {
	voteResult := map[string]uint64{}
	data, err := scs.GetData(sortedlistkey)
	if err != nil {
		return nil, err
	}
	if len(data) != 0 {
		dec := gob.NewDecoder(bytes.NewBuffer(data))
		var voteList types.VoteList
		err = dec.Decode(&voteList)
		if err != nil {
			return nil, err
		}
		for _, v := range voteList.GetVotes() {
			voteResult[base58.Encode(v.Candidate)] = v.Amount
		}
	}
	return &voteResult, nil
}

func InitVoteResult(scs *state.ContractState, voteResult *map[string]uint64) error {
	if voteResult == nil {
		return errors.New("Invalid argument : voteReult should not nil")
	}
	return syncVoteResult(scs, voteResult)
}

func syncVoteResult(scs *state.ContractState, voteResult *map[string]uint64) error {
	var voteList types.VoteList
	for k, v := range *voteResult {
		c, _ := base58.Decode(k)
		vote := &types.Vote{
			Candidate: c,
			Amount:    v,
		}
		voteList.Votes = append(voteList.Votes, vote)
	}
	sort.Sort(sort.Reverse(voteList))
	//logger.Info().Msgf("VOTE set list %v", voteList.Votes)
	data, err := common.GobEncode(voteList)
	if err != nil {
		return err
	}
	return scs.SetData(sortedlistkey, data)
}

func GetVoteResult(scs *state.ContractState) (*types.VoteList, error) {
	data, err := scs.GetData(sortedlistkey)
	if err != nil {
		return nil, err
	}

	voteList := &types.VoteList{}
	err = common.GobDecode(data, voteList)
	if err != nil {
		return nil, err
	}
	return voteList, nil
}
