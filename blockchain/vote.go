/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sort"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const limitDuration = 23
const sortedlist = "sortedlist"

func executeVoteTx(txBody *types.TxBody, senderState *types.State,
	receiverState *types.State, scs *state.ContractState, blockNo types.BlockNo) error {
	voteCmd := txBody.GetPayload()[0]
	if voteCmd == 'v' { //staking, vote
		if senderState.Balance < txBody.Amount {
			return errors.New("not enough balance")
		}
		voting, _, err := getVote(scs, txBody.Account, txBody.Payload[1:])
		if err != nil {
			return err
		}
		err = setVote(scs, txBody.Account, txBody.Payload[1:], voting+txBody.Amount, blockNo)
		if err != nil {
			return err
		}
		senderState.Balance = senderState.Balance - txBody.Amount
		//update candidate total
		err = updateVoteResult(scs, txBody.Payload[1:], (int64)(txBody.Amount), blockNo)
		if err != nil {
			return err
		}
	} else if voteCmd == 'r' { //unstaking, revert
		voting, blockNo, err := getVote(scs, txBody.Account, txBody.Payload[1:])
		if blockNo < limitDuration+blockNo { //TODO : fix it proper
			return errors.New("less time has passed")
		}
		err = setVote(scs, txBody.Account, txBody.Payload[1:], 0, blockNo)
		if err != nil {
			return err
		}
		err = updateVoteResult(scs, txBody.Payload[1:], -(int64)(voting), blockNo)
		if err != nil {
			return err
		}
		senderState.Balance = senderState.Balance + voting
	}
	return nil
}

func getVote(scs *state.ContractState, voter []byte, to []byte) (uint64, uint64, error) {
	key := append(voter, to...)
	return getVoteData(scs, key)
}

func setVote(scs *state.ContractState, voter []byte, to []byte, amount uint64, blockNo uint64) error {
	//update personal voting infomation
	key := append(voter, to...)
	err := setVoteData(scs, key, amount, blockNo)
	if err != nil {
		return err
	}
	return nil
}

func updateVoteResult(scs *state.ContractState, key []byte, amount int64, blockNo uint64) error {
	var voteList types.VoteList
	var tmp []*types.Vote
	voteList.Votes = tmp
	isInList := false

	votes, _, err := getVoteData(scs, key)
	if err != nil {
		return err
	}
	//logger.Info().Uint64("votes", votes).Msg("VOTE updateVote")
	err = setVoteData(scs, key, (uint64)((int64)(votes)+amount), blockNo)
	if err != nil {
		return err
	}
	data, err := scs.GetData([]byte(sortedlist))
	if err != nil {
		return err
	}
	//logger.Info().Str("key", util.EncodeB64(key)).Msg("VOTE updateVote")
	//TODO: fix hardcorded length, '39' length of peer id now
	for offset := 0; offset < len(data); offset += (39 + 8) {
		vote := &types.Vote{
			Candidate: data[offset : offset+39],
			Amount:    binary.LittleEndian.Uint64(data[offset+39 : offset+39+8]),
		}
		if bytes.Equal(key, vote.Candidate) {
			if votes != vote.Amount {
				panic("voting data crashed")
			}
			vote.Amount = (uint64)((int64)(vote.Amount) + amount)
			if vote.Amount != 0 {
				voteList.Votes = append(voteList.Votes, vote)
			}
			isInList = true
		} else {
			voteList.Votes = append(voteList.Votes, vote)
		}
		//logger.Info().Str("key in list", util.EncodeB64(vote.Candidate)).Msg("VOTE updateVote")
		//logger.Info().Uint64("votelist.amount", vote.Amount).Msg("VOTE updateVote")
	}
	if !isInList {
		vote := &types.Vote{
			Candidate: key,
			Amount:    (uint64)(amount),
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
	return scs.SetData([]byte(sortedlist), buf)
}

func setVoteData(scs *state.ContractState, key []byte, balance uint64, blockNo uint64) error {
	v := make([]byte, 16)
	binary.LittleEndian.PutUint64(v, balance)
	binary.LittleEndian.PutUint64(v[8:], blockNo) //TODO:change to block no
	//logger.Info().Str("key", util.EncodeB64(key)).Msg("VOTE setVote")
	//logger.Info().Uint64("balance", balance).Uint64("blockNo", blockNo).Msg("VOTE setVote")
	return scs.SetData(key, v)
}

func getVoteData(scs *state.ContractState, key []byte) (uint64, uint64, error) {
	data, err := scs.GetData(key)
	if err != nil {
		return 0, 0, err
	}
	var balance uint64
	var blockNo uint64
	if cap(data) == 0 {
		balance = 0
		blockNo = 0
	} else if cap(data) >= 8 {
		balance = binary.LittleEndian.Uint64(data[:8])
		blockNo = 0
		if cap(data) >= 16 {
			blockNo = binary.LittleEndian.Uint64(data[8:16])
		}
	}
	//logger.Info().Str("key", util.EncodeB64(key)).Msg("VOTE getVote")
	//logger.Info().Uint64("balance", balance).Uint64("blockNo", blockNo).Msg("VOTE getVote")
	return balance, blockNo, nil
}

func getVoteResult(scs *state.ContractState, n int) (*types.VoteList, error) {
	var voteList types.VoteList
	data, err := scs.GetData([]byte(sortedlist))
	if err != nil {
		return nil, err
	}
	var tmp []*types.Vote
	voteList.Votes = tmp
	i := 0
	for offset := 0; offset < len(data) && i < n; offset += (39 + 8) {
		vote := &types.Vote{
			Candidate: data[offset : offset+39],
			Amount:    binary.LittleEndian.Uint64(data[offset+39 : offset+39+8]),
		}
		voteList.Votes = append(voteList.Votes, vote)
		i++
	}
	//logger.Info().Msgf("VOTE get %v", voteList.Votes)
	return &voteList, nil
}

func (cs *ChainService) getVotes(n int) (*types.VoteList, error) {
	scs, err := cs.sdb.OpenContractStateAccount(types.ToAccountID([]byte(aergobp)))
	if err != nil {
		return nil, err
	}
	return getVoteResult(scs, n)
}
