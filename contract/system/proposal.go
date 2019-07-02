package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const (
	bpCount sysParamIndex = iota // BP count
	numBP                        // BP count
	sysParamMax
)

const proposalPrefixKey = "proposal" //aergo proposal format

var proposalListKey = []byte("proposallist")

//go:generate stringer -type=sysParamIndex
type sysParamIndex int

func (i sysParamIndex) ID() string {
	return strings.ToUpper(i.String())
}

func GetVotingIssues() []types.VotingIssue {
	vi := make([]types.VotingIssue, sysParamMax)
	for i := bpCount; i < sysParamMax; i++ {
		vi[int(i)] = i
	}
	return vi
}

type whereToVotes = [][]byte

type Proposal struct {
	ID             string
	Description    string
	Blockfrom      uint64
	Blockto        uint64
	MultipleChoice uint32
	Candidates     []string
}

func (a *Proposal) GetKey() []byte {
	return []byte(proposalPrefixKey + "\\" + strings.ToUpper(a.ID))
}

func GenProposalKey(id string) []byte {
	return []byte(proposalPrefixKey + "\\" + strings.ToUpper(id))
}

func ProposalIDfromKey(key []byte) string {
	return strings.Replace(string(key), proposalPrefixKey+"\\", "", 1)
}

type proposalCmd struct {
	*SystemContext
	amount *big.Int
}

func newProposalCmd(ctx *SystemContext) (sysCmd, error) {
	return &proposalCmd{SystemContext: ctx, amount: ctx.txBody.GetAmountBigInt()}, nil
}

func (c *proposalCmd) run() (*types.Event, error) {
	var (
		scs      = c.scs
		proposal = c.Proposal
		sender   = c.Sender
		receiver = c.Receiver
		amount   = c.amount
	)

	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	if err := setProposal(scs, proposal); err != nil {
		return nil, err
	}
	log, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       c.op.ID(),
		JsonArgs: `{"who":"` +
			types.EncodeAddress(sender.ID()) +
			`", "Proposal":` + string(log) + `}`,
	}, nil

}

//getProposal find proposal using id
func getProposal(scs *state.ContractState, id string) (*Proposal, error) {
	dataKey := GenProposalKey(id)
	data, err := scs.GetData([]byte(dataKey))
	if err != nil {
		return nil, fmt.Errorf("could not get proposal from contract state DB : %s", id)
	}
	return deserializeProposal(data), nil
}

func setProposal(scs *state.ContractState, proposal *Proposal) error {
	return scs.SetData(proposal.GetKey(), serializeProposal(proposal))
}

func serializeProposal(proposal *Proposal) []byte {
	data, err := json.Marshal(proposal)
	if err != nil {
		panic("could not marshal proposal")
	}
	return data
}

func deserializeProposal(data []byte) *Proposal {
	var proposal Proposal
	if err := json.Unmarshal(data, &proposal); err != nil {
		return nil
	}
	return &proposal
}

func getProposalHistory(scs *state.ContractState, address []byte) whereToVotes {
	key := append(proposalListKey, address...)
	return _getProposalHistory(scs, key)
}
func _getProposalHistory(scs *state.ContractState, key []byte) whereToVotes {
	data, err := scs.GetData(key)
	if err != nil {
		panic("could not get proposal history in contract state db")
	}
	if len(data) == 0 { //never vote before
		return nil
	}
	return deserializeProposalHistory(data)
}

func addProposalHistory(scs *state.ContractState, address []byte, proposal *Proposal) error {

	// address: sender's address
	key := append(proposalListKey, address...)

	// scs corresponds to aergo.system or aergo.name. proposalHistory includes
	// ` delimited proposal keys, which are the prefixed IDs of the
	// proposals. Its purpose is to memorize which kind of voting has been
	// cast.
	proposalHistory := _getProposalHistory(scs, key)
	proposalHistory = append(proposalHistory, proposal.GetKey())

	//unique
	filter := make(map[string]bool)
	var result whereToVotes
	for _, entryBytes := range proposalHistory {
		entry := string(entryBytes)
		if _, exist := filter[entry]; !exist {
			filter[entry] = true
			result = append(result, entryBytes)
		}
	}

	return scs.SetData(key, serializeProposalHistory(result))
}

func deserializeProposalHistory(data []byte) whereToVotes {
	return bytes.Split(data, []byte("`"))
}

func serializeProposalHistory(wtv whereToVotes) []byte {
	var data []byte
	for i, w := range wtv {
		if i != 0 {
			data = append(data, '`')
		}
		data = append(data, w...)
	}
	return data
}

func isValidID(id string) bool {
	for i := sysParamIndex(0); i < sysParamMax; i++ {
		if strings.ToUpper(id) == i.ID() {
			return true
		}
	}
	return false
}
