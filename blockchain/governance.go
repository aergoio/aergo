/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"errors"
	"sort"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
)

const minimum = 1000

func (cs *ChainService) processGovernanceTx(dbtx *db.Transaction, bs *state.BlockState, txBody *types.TxBody) error {
	if txBody.Amount < minimum {
		return errors.New("too small amount to influence")
	}
	governanceCmd := string(txBody.GetRecipient())

	var err error
	switch governanceCmd {
	case "aergo.bp":
		err = cs.processVoteTx(dbtx, bs, txBody)
	default:
		logger.Warn().Str("governanceCmd", governanceCmd).Msg("receive unknown cmd")
	}
	return err
}

func (cs *ChainService) processVoteTx(dbtx *db.Transaction, bs *state.BlockState, txBody *types.TxBody) error {
	senderID := types.ToAccountID(txBody.Account)
	senderState, err := cs.sdb.GetBlockAccountClone(bs, senderID)
	if err != nil {
		return err
	}
	senderChange := types.Clone(*senderState).(types.State)

	voter := types.EncodeB64(txBody.Account)
	c, err := peer.IDFromBytes(txBody.Payload[1:])
	if err != nil {
		return err
	}
	to := peer.IDB58Encode(c)
	if txBody.Payload[0] == 'v' { //staking, vote
		if senderChange.Balance < txBody.Amount {
			return errors.New("not enough balance")
		}
		senderChange.Balance = senderState.Balance - txBody.Amount
		cs.putVote(voter, to, int64(txBody.Amount))
		senderChange.Nonce = txBody.Nonce
		bs.PutAccount(senderID, senderState, &senderChange)

	} else if txBody.Payload[0] == 'r' { //unstaking, revert
		//TODO: check valid candidate, voter, amount from state db
		if cs.getVote(voter, to) < txBody.Amount {
			return errors.New("not enough staking balance")
		}
		senderChange.Balance = senderState.Balance + txBody.Amount
		bs.PutAccount(senderID, senderState, &senderChange)
		//TODO: update candidate, voter, amount to state db
		cs.putVote(voter, to, -int64(txBody.Amount))
	}
	return nil
}

func (cs *ChainService) putVote(voter string, to string, amount int64) {
	//TODO: update candidate, voter, amount to state db
	entry, ok := cs.voters[voter]
	if !ok {
		entry = make(map[string][]uint64)
		cs.voters[voter] = entry
	}
	amountBlockNo, ok := cs.voters[voter][to]
	if !ok {
		amountBlockNo = make([]uint64, 2)
		cs.voters[voter][to] = amountBlockNo
	}

	entry[to][0] = uint64(int64(entry[to][0]) + amount)
	cs.votes[to] = uint64(int64(cs.votes[to]) + amount)
}

func (cs *ChainService) getVote(voter string, to string) uint64 {
	return cs.voters[voter][to][0]
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
