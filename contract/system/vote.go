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
	"strings"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
)

const (
	PeerIDLength = 39
	VotingDelay  = 60 * 60 * 24 //block interval
	//VotingDelay = 10 //block interval
)

var (
	votingCatalog  []types.VotingIssue
	lastBpCount    int
	defaultVoteKey = []byte(types.OpvoteBP.ID())
)

func init() {
	initVotingCatalog()
}

func initVotingCatalog() {
	votingCatalog = make([]types.VotingIssue, 0)

	fuse := func(issues []types.VotingIssue) {
		votingCatalog = append(votingCatalog, issues...)
	}

	fuse(types.GetVotingIssues())
	fuse(GetVotingIssues())
}

func GetVotingCatalog() []types.VotingIssue {
	return votingCatalog
}

type vprCmd struct {
	*SystemContext
	voteResult *VoteResult

	add func(v *types.Vote) error
	sub func(v *types.Vote) error
}

func newVprCmd(ctx *SystemContext, vr *VoteResult) *vprCmd {
	cmd := &vprCmd{SystemContext: ctx, voteResult: vr}

	if vprLogger.IsDebugEnabled() {
		vprLogger.Debug().
			Int32("block version", ctx.BlockInfo.ForkVersion).
			Msg("create new voting power table command")
	}

	if ctx.BlockInfo.ForkVersion < 2 {
		cmd.add = func(v *types.Vote) error {
			return cmd.voteResult.AddVote(v)
		}
		cmd.sub = func(v *types.Vote) error {
			return cmd.voteResult.SubVote(v)
		}
	} else {
		cmd.add = func(v *types.Vote) error {
			cmd.addVpr(v) // calculate voting power rank with exception handling
			return cmd.voteResult.AddVote(v)
		}
		cmd.sub = func(v *types.Vote) error {
			cmd.subVpr(v) // calculate voting power rank with exception handling
			return cmd.voteResult.SubVote(v)
		}
	}

	return cmd
}

func (c *vprCmd) subVpr(v *types.Vote) {
	no := c.BlockInfo.No
	aid := c.Sender.AccountID()

	// Handle exception 1. multisig contract ( AmhNcvE7RR84xoRzYNyATnwZR2JXaC5ut7neu89R13aj1b4eUxKp )
	if aid.String() == "A9zXKkooeGYAZC5ReCcgeg4ddsvMHAy2ivUafXhrnzpj" {
		votingPowerRank.sub(statedb.EmptyAccountID, c.Sender.ID(), v.GetAmountBigInt())
		return
	}

	// Handle exception 2. reproduce vpr calculation
	// When block is reverted, votingPowerRank is not reverted and calculated three times.
	if aid.String() == "36t2u7Q31HmEbkkYZng7DHNm3xepxHKUfgGrAXNA8pMW" && no == 138015125 {
		for i := 0; i < 3; i++ {
			votingPowerRank.sub(aid, c.Sender.ID(), v.GetAmountBigInt())
		}
		return
	}

	// normal case
	votingPowerRank.sub(aid, c.Sender.ID(), v.GetAmountBigInt())
}

func (c *vprCmd) addVpr(v *types.Vote) {
	no := c.BlockInfo.No
	aid := c.Sender.AccountID()

	// Handle exception 1. multisig contract ( AmhNcvE7RR84xoRzYNyATnwZR2JXaC5ut7neu89R13aj1b4eUxKp )
	if aid.String() == "A9zXKkooeGYAZC5ReCcgeg4ddsvMHAy2ivUafXhrnzpj" {
		votingPowerRank.add(statedb.EmptyAccountID, c.Sender.ID(), v.GetAmountBigInt())
		return
	}

	// Handle exception 2. reproduce vpr calculation
	// When block is reverted, votingPowerRank is not reverted and calculated three times.
	if aid.String() == "36t2u7Q31HmEbkkYZng7DHNm3xepxHKUfgGrAXNA8pMW" && no == 138015125 {
		for i := 0; i < 3; i++ {
			votingPowerRank.add(aid, c.Sender.ID(), v.GetAmountBigInt())
		}
		return
	}

	// normal case
	votingPowerRank.add(aid, c.Sender.ID(), v.GetAmountBigInt())
}

type voteCmd struct {
	*vprCmd

	issue     []byte
	args      []byte
	candidate []byte

	newVote *types.Vote
}

func newVoteCmd(ctx *SystemContext) (sysCmd, error) {
	var (
		scs = ctx.scs

		err error
	)

	cmd := &voteCmd{}

	if ctx.Proposal != nil {
		cmd.issue = ctx.Proposal.GetKey()
		cmd.candidate, err = json.Marshal(ctx.Call.Args[1:]) //[0] is name
		if err != nil {
			return nil, err
		}
		//for event. voteDAO allow only one candidate. it should be validate before.
		voteID := ctx.Call.Args[0].(string)
		cmd.args = []byte(`"` + strings.ToUpper(voteID) + `", {"_bignum":"` + ctx.Call.Args[1].(string) + `"}`)
	} else {
		cmd.issue = []byte(ctx.op.ID())
		cmd.args, err = json.Marshal(ctx.Call.Args)
		if err != nil {
			return nil, err
		}
		for _, v := range ctx.Call.Args {
			candidate, _ := base58.Decode(v.(string))
			cmd.candidate = append(cmd.candidate, candidate...)
		}
	}

	// The variable args is a JSON bytes. It is used as vote.candidate for the
	// proposal based voting, while just as an event output for BP election.
	staked := ctx.Staked
	// Update the block number when the last action is conducted (voting,
	// staking etc). Two consecutive votings must be seperated by the time
	// corresponding to VotingDeley (currently 24h). This time limit is check
	// against this block number (Staking.When). Due to this, the Staking value
	// on the state DB must be updated even for voting.
	staked.SetWhen(ctx.BlockInfo.No)

	if staked.GetAmountBigInt().Cmp(new(big.Int).SetUint64(0)) == 0 {
		return nil, types.ErrMustStakeBeforeVote
	}

	cmd.newVote = &types.Vote{
		Candidate: cmd.candidate,
		Amount:    staked.GetAmount(),
	}

	voteResult, err := loadVoteResult(scs, cmd.issue)
	if err != nil {
		return nil, err
	}

	cmd.vprCmd = newVprCmd(ctx, voteResult)

	return cmd, err
}

func (c *voteCmd) run() (*types.Event, error) {
	// To update Staking.When field (not Staking.Amount).
	if err := c.updateStaking(); err != nil {
		return nil, err
	}

	if err := c.updateVote(); err != nil {
		return nil, err
	}

	if err := c.updateVoteResult(); err != nil {
		return nil, err
	}

	jsonArgs := ""
	if c.SystemContext.BlockInfo.ForkVersion < 2 {
		jsonArgs = `{"who":"` + types.EncodeAddress(c.txBody.Account) + `", "vote":` + string(c.args) + `}`
	} else {
		jsonArgs = `["` + types.EncodeAddress(c.txBody.Account) + `", ` + string(c.args) + `]`
	}
	return &types.Event{
		ContractAddress: c.Receiver.ID(),
		EventIdx:        0,
		EventName:       c.op.ID(),
		JsonArgs:        jsonArgs,
	}, nil
}

// Update the sender's voting record.
func (c *voteCmd) updateVote() error {
	return setVote(c.scs, c.issue, c.Sender.ID(), c.newVote)
}

// Apply the new voting to the voting statistics on the (system) contract
// storage.
func (c *voteCmd) updateVoteResult() error {
	if err := c.sub(c.Vote); err != nil {
		return err
	}

	if err := c.add(c.newVote); err != nil {
		return err
	}

	if vprLogger.IsDebugEnabled() {
		vprLogger.Debug().
			Str("sub", c.Vote.GetAmountBigInt().String()).
			Str("add", c.Vote.GetAmountBigInt().String()).
			Msg("update vote result")
	}

	return c.voteResult.Sync()
}

func refreshAllVote(context *SystemContext) error {
	var (
		scs          = context.scs
		account      = context.Sender.ID()
		staked       = context.Staked
		stakedAmount = new(big.Int).SetBytes(staked.Amount)
	)

	for _, i := range GetVotingCatalog() {
		key := i.Key()

		oldvote, err := getVote(scs, key, account)
		if err != nil {
			return err
		}
		if oldvote.Amount == nil ||
			new(big.Int).SetBytes(oldvote.Amount).Cmp(stakedAmount) <= 0 {
			continue
		}
		if types.OpvoteBP.ID() != i.ID() {
			proposal, err := getProposal(i.ID())
			if err != nil {
				return err
			}
			if proposal != nil && proposal.Blockto != 0 && proposal.Blockto < context.BlockInfo.No {
				continue
			}
		}
		voteResult, err := loadVoteResult(scs, key)
		if err != nil {
			return err
		}

		cmd := newVprCmd(context, voteResult)

		if err = cmd.sub(oldvote); err != nil {
			return err
		}
		oldvote.Amount = staked.GetAmount()
		if err = setVote(scs, key, account, oldvote); err != nil {
			return err
		}
		if err = cmd.add(oldvote); err != nil {
			return err
		}
		if err = voteResult.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// GetVote return amount, to, err.
func GetVote(scs *statedb.ContractState, voter []byte, issue []byte) (*types.Vote, error) {
	return getVote(scs, issue, voter)
}

func getVote(scs *statedb.ContractState, key, voter []byte) (*types.Vote, error) {
	data, err := scs.GetData(dbkey.SystemVote(key, voter))
	if err != nil {
		return nil, err
	}

	if len(data) != 0 {
		if bytes.Equal(key, defaultVoteKey) {
			return deserializeVote(data), nil
		} else {
			return deserializeVoteEx(data), nil
		}
	}

	return &types.Vote{}, nil
}

func setVote(scs *statedb.ContractState, key, voter []byte, vote *types.Vote) error {
	if bytes.Equal(key, defaultVoteKey) {
		return scs.SetData(dbkey.SystemVote(key, voter), serializeVote(vote))
	} else {
		return scs.SetData(dbkey.SystemVote(key, voter), serializeVoteEx(vote))
	}
}

// BuildOrderedCandidates returns a candidate list ordered by votes.xs
func BuildOrderedCandidates(vote map[string]*big.Int) []string {
	// TODO: cleanup
	voteResult := newVoteResult(defaultVoteKey, nil)
	voteResult.rmap = vote
	l := voteResult.buildVoteList()
	bps := make([]string, 0, len(l.Votes))
	for _, v := range l.Votes {
		bp := base58.Encode(v.Candidate)
		bps = append(bps, bp)
	}
	return bps
}

// GetVoteResult returns the top n voting result from the system account state.
func GetVoteResult(scs *statedb.ContractState, id []byte, n int) (*types.VoteList, error) {
	if !bytes.Equal(id, defaultVoteKey) {
		id = GenProposalKey(string(id))
	}
	return getVoteResult(scs, id, n)
}

// GetRankers returns the IDs of the top n rankers.
func GetRankers(scs *statedb.ContractState) ([]string, error) {
	n := GetBpCount()

	vl, err := GetVoteResult(scs, defaultVoteKey, n)
	if err != nil {
		return nil, err
	}

	bps := make([]string, 0, n)
	for _, v := range vl.Votes {
		bps = append(bps, base58.Encode(v.Candidate))
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
