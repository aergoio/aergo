/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"encoding/binary"
	"sort"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
)

var votingkey = []byte("voting")
var totalkey = []byte("totalvote")
var sortedlistkey = []byte("sortedlist")

const peerIDLength = 39
const votingDelay = 5

func voting(txBody *types.TxBody, scs *state.ContractState, blockNo types.BlockNo) error {
	old, when, candidates, err := getVote(scs, txBody.Account)
	if err != nil {
		return err
	}
	if when+votingDelay > blockNo {
		logger.Debug().Uint64("when", when).Uint64("blockNo", blockNo).Msg("remain voting delay")
		return types.ErrLessTimeHasPassed
	}
	staked, when, err := getStaking(scs, txBody.Account)
	if err != nil {
		return err
	}
	if when+votingDelay > blockNo {
		logger.Debug().Uint64("when", when).Uint64("blockNo", blockNo).Msg("remain voting delay")
		return types.ErrLessTimeHasPassed
	}
	err = setStaking(scs, txBody.Account, staked, blockNo)
	if err != nil {
		return err
	}

	voteResult, err := loadVoteResult(scs)
	for offset := 0; offset < len(candidates); offset += peerIDLength {
		key := candidates[offset : offset+peerIDLength]
		(*voteResult)[base58.Encode(key)] -= old
	}

	if txBody.Payload[0] != 'v' { //called from unstaking
		err = setVote(scs, txBody.Account, candidates, staked, blockNo)
		if err != nil {
			return err
		}
		for offset := 0; offset < len(candidates); offset += peerIDLength {
			key := candidates[offset : offset+peerIDLength]
			(*voteResult)[base58.Encode(key)] += staked
		}
	} else {
		if staked == 0 {
			return types.ErrMustStakeBeforeVote
		}
		err = setVote(scs, txBody.Account, txBody.Payload[1:], staked, blockNo)
		if err != nil {
			return err
		}
		for offset := 0; offset < len(txBody.Payload[1:]); offset += peerIDLength {
			key := txBody.Payload[offset+1 : offset+peerIDLength+1]
			(*voteResult)[base58.Encode(key)] += staked
		}
	}

	err = syncVoteResult(scs, voteResult)
	if err != nil {
		return err
	}
	return nil
}

//getVote return amount, when, to, err
func getVote(scs *state.ContractState, voter []byte) (uint64, uint64, []byte, error) {
	key := append(votingkey, voter...)
	data, err := scs.GetData(key)
	if err != nil {
		return 0, 0, nil, err
	}
	if len(data) == 0 {
		return 0, 0, nil, nil
	}
	return binary.LittleEndian.Uint64(data[:8]),
		binary.LittleEndian.Uint64(data[8:16]), data[16:], nil
}

func setVote(scs *state.ContractState, voter []byte, to []byte, amount uint64,
	blockNo uint64) error {
	key := append(votingkey, voter...)
	//TODO : change to key iteration
	var data []byte
	whenAmount := make([]byte, 16)
	binary.LittleEndian.PutUint64(whenAmount, amount)
	binary.LittleEndian.PutUint64(whenAmount[8:], blockNo)
	data = append(data, whenAmount...)
	data = append(data, to...)
	return scs.SetData(key, data)
}

func loadVoteResult(scs *state.ContractState) (*map[string]uint64, error) {
	voteResult := map[string]uint64{}
	data, err := scs.GetData(sortedlistkey)
	if err != nil {
		return nil, err
	}
	for offset := 0; offset < len(data); offset += (peerIDLength + 8) {
		value := binary.LittleEndian.Uint64(data[offset+peerIDLength : offset+peerIDLength+8])
		voteResult[base58.Encode(data[offset:offset+peerIDLength])] = value
	}
	return &voteResult, nil
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
	var buf []byte
	vbuf := make([]byte, 8)
	for _, v := range voteList.Votes {
		votes := v.Candidate
		binary.LittleEndian.PutUint64(vbuf, v.Amount)
		buf = append(buf, votes...)
		buf = append(buf, vbuf...)
	}
	return scs.SetData(sortedlistkey, buf)
}

func updateVoteResult(scs *state.ContractState, candidates []byte, amount uint64, plus bool) error {
	voteResult, err := loadVoteResult(scs)
	total := make([]byte, 8)
	for offset := 0; offset < len(candidates); offset += peerIDLength {
		key := candidates[offset : offset+peerIDLength]
		current := (*voteResult)[base58.Encode(key)]
		if plus {
			(*voteResult)[base58.Encode(key)] = current + amount
			binary.LittleEndian.PutUint64(total, current+amount)
			err = scs.SetData(key, total)
		} else /* minus */ {
			if current > amount {
				(*voteResult)[base58.Encode(key)] = current - amount
				binary.LittleEndian.PutUint64(total, current-amount)
				err = scs.SetData(key, total)
			} else {
				(*voteResult)[base58.Encode(key)] = 0
				binary.LittleEndian.PutUint64(total, 0)
				err = scs.SetData(key, total)
			}
		}
	}
	if err != nil {
		return err
	}
	var voteList types.VoteList
	for k, v := range *voteResult {
		c, _ := base58.Decode(k)
		vote := &types.Vote{
			Candidate: c,
			Amount:    v,
		}
		if vote.Amount > 0 {
			voteList.Votes = append(voteList.Votes, vote)
		}
	}
	sort.Sort(sort.Reverse(voteList))
	//logger.Info().Msgf("VOTE set list %v", voteList.Votes)
	var buf []byte
	vbuf := make([]byte, 8)
	for _, v := range voteList.Votes {
		votes := v.Candidate
		binary.LittleEndian.PutUint64(vbuf, v.Amount)
		buf = append(buf, votes...)
		buf = append(buf, vbuf...)
	}
	return scs.SetData(sortedlistkey, buf)
}

func cleanupVoting(scs *state.ContractState, who []byte, amount uint64,
	blockNo types.BlockNo, remainStaking bool) error {
	//clean up voting
	_, when, candidates, err := getVote(scs, who)
	if err != nil {
		return err
	}
	if blockNo < when+votingDelay {
		return types.ErrLessTimeHasPassed
	}
	if !remainStaking {
		err = setVote(scs, who, nil, 0, blockNo)
		if err != nil {
			return err
		}
	}
	return updateVoteResult(scs, candidates, amount, false)
}

func getVoteResult(scs *state.ContractState, n int) (*types.VoteList, error) {
	var voteList types.VoteList
	data, err := scs.GetData(sortedlistkey)
	if err != nil {
		return nil, err
	}
	var tmp []*types.Vote
	voteList.Votes = tmp
	i := 0
	for offset := 0; offset < len(data) && i < n; offset += (peerIDLength + 8) {
		vote := &types.Vote{
			Candidate: data[offset : offset+peerIDLength],
			Amount:    binary.LittleEndian.Uint64(data[offset+peerIDLength : offset+peerIDLength+8]),
		}
		voteList.Votes = append(voteList.Votes, vote)
		i++
	}
	//logger.Info().Msgf("VOTE get %v", voteList.Votes)
	return &voteList, nil
}

func (cs *ChainService) getVotes(n int) (*types.VoteList, error) {
	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	return getVoteResult(scs, n)
}

func (cs *ChainService) getVote(addr []byte) (*types.VoteList, error) {
	scs, err := cs.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte(types.AergoSystem)))
	if err != nil {
		return nil, err
	}
	var voteList types.VoteList
	var tmp []*types.Vote
	voteList.Votes = tmp
	amount, _, to, err := getVote(scs, addr)
	if err != nil {
		return nil, err
	}
	for offset := 0; offset < len(to); offset += peerIDLength {
		vote := &types.Vote{
			Candidate: to[offset : offset+peerIDLength],
			Amount:    amount,
		}
		voteList.Votes = append(voteList.Votes, vote)
	}
	return &voteList, nil
}
