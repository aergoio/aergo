package blockchain

import (
	"encoding/binary"
	"errors"
	"sort"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/mr-tron/base58/base58"
)

const limitDuration = 23

func (cs *ChainService) processVoteTx(bs *state.BlockState, scs *state.ContractState, txBody *types.TxBody, block *types.Block) error {
	senderID := types.ToAccountID(txBody.Account)
	senderState, err := cs.sdb.GetBlockAccountClone(bs, senderID)
	if err != nil {
		return err
	}
	senderChange := types.Clone(*senderState).(types.State)
	voteCmd := txBody.GetPayload()[0]
	if voteCmd == 'v' { //staking, vote
		if senderChange.Balance < txBody.Amount {
			return errors.New("not enough balance")
		}
		voting, _, err := cs.getVote(scs, txBody.Account, txBody.Payload[1:])
		if err != nil {
			return err
		}
		err = cs.setVote(scs, txBody.Account, txBody.Payload[1:], voting+txBody.Amount, block.GetHeader().GetBlockNo())
		if err != nil {
			return err
		}
		senderChange.Balance = senderState.Balance - txBody.Amount
		senderChange.Nonce = txBody.Nonce
		bs.PutAccount(senderID, senderState, &senderChange)
	} else if voteCmd == 'r' { //unstaking, revert
		voting, blockNo, err := cs.getVote(scs, txBody.Account, txBody.Recipient)
		if blockNo < limitDuration { //TODO : fix it proper
			return errors.New("less time has passed")
		}
		err = cs.setVote(scs, txBody.Account, txBody.Payload[1:], 0, block.GetHeader().GetBlockNo())
		if err != nil {
			return err
		}
		senderChange.Balance = senderState.Balance + voting
		bs.PutAccount(senderID, senderState, &senderChange)
	}

	return nil
}

func (cs *ChainService) loadVotes(scs *state.ContractState) error {
	data, err := scs.GetData([]byte(aergobp))
	if err != nil {
		return err
	}
	for i := 0; i < len(data); i += (39 + 8) {
		cs.votes[base58.Encode(data[i:i+39])] = binary.LittleEndian.Uint64(data[i+39 : i+39+8])
	}
	return nil
}

func (cs *ChainService) updateVoteCache(voter string, to string, amount uint64) {
	cs.votes[to] = amount
}

func (cs *ChainService) syncVoteCache(scs *state.ContractState) error {
	var buf []byte
	vbuf := make([]byte, 8)

	for k, v := range cs.votes {
		binary.LittleEndian.PutUint64(vbuf, v)
		votes, _ := base58.Decode(k)
		votes = append(votes, vbuf...)
		buf = append(buf, votes...)
	}
	return scs.SetData([]byte(aergobp), buf)
}

func (cs *ChainService) getVote(scs *state.ContractState, voter []byte, to []byte) (uint64, uint64, error) {
	key := append(voter, to...)
	return getVoteData(scs, key)
}

func (cs *ChainService) setVote(scs *state.ContractState, voter []byte, to []byte, amount uint64, blockNo uint64) error {
	//update personal voting infomation
	key := append(voter, to...)
	err := setVoteData(scs, key, amount, 0) //TODO : 0 to block number
	if err != nil {
		return err
	}

	//update candidate total
	err = setVoteData(scs, to, amount, 0)
	if err != nil {
		return err
	}

	peerID, err := peer.IDFromBytes(to)
	if err != nil {
		return err
	}
	voterB58 := base58.Encode(voter)
	toB58 := peer.IDB58Encode(peerID)

	cs.updateVoteCache(voterB58, toB58, amount)
	cs.syncVoteCache(scs)
	return nil
}

func setVoteData(scs *state.ContractState, key []byte, balance uint64, blockNo uint64) error {
	v := make([]byte, 16)
	binary.LittleEndian.PutUint64(v, balance)
	binary.LittleEndian.PutUint64(v[:8], blockNo) //TODO:change to block no
	err := scs.SetData(key, v)
	if err != nil {
		return err
	}
	return nil
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
	return balance, blockNo, nil
}

func (cs *ChainService) getVotes(n int) types.VoteList {
	var ret types.VoteList
	tmp := make([]*types.Vote, len(cs.votes))
	ret.Votes = tmp

	i := 0
	for k, v := range cs.votes {
		c := types.DecodeB58(k)
		ret.Votes[i] = &types.Vote{Candidate: c, Amount: v}
		i++
	}
	sort.Sort(sort.Reverse(ret))
	if n < len(cs.votes) {
		ret.Votes = ret.Votes[:n]
	}
	return ret
}
