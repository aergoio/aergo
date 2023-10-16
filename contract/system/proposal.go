package system

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/types"
)

const proposalPrefixKey = "proposal" //aergo proposal format

func (i sysParamIndex) ID() string {
	return strings.ToUpper(i.String())
}

func (i sysParamIndex) Key() []byte {
	return GenProposalKey(i.String())
}

func GetVotingIssues() []types.VotingIssue {
	vi := make([]types.VotingIssue, sysParamMax)
	for i := sysParamIndex(0); i < sysParamMax; i++ {
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
	Default        *big.Int
}

var SystemProposal = map[string]*Proposal{
	bpCount.ID(): {
		ID:             bpCount.ID(),
		Description:    "",
		Blockfrom:      0,
		Blockto:        0,
		MultipleChoice: 1,
		Candidates:     nil,
	},
	stakingMin.ID(): {
		ID:             stakingMin.ID(),
		Description:    "",
		Blockfrom:      0,
		Blockto:        0,
		MultipleChoice: 1,
		Candidates:     nil,
	},
	gasPrice.ID(): {
		ID:             gasPrice.ID(),
		Description:    "",
		Blockfrom:      0,
		Blockto:        0,
		MultipleChoice: 1,
		Candidates:     nil,
	},
	namePrice.ID(): {
		ID:             namePrice.ID(),
		Description:    "",
		Blockfrom:      0,
		Blockto:        0,
		MultipleChoice: 1,
		Candidates:     nil,
	},
}

func (a *Proposal) GetKey() []byte {
	return []byte(strings.ToUpper(a.ID))
}

func GenProposalKey(id string) []byte {
	return []byte(strings.ToUpper(id))
}

/*
func ProposalIDfromKey(key []byte) string {
	return strings.Replace(string(key), proposalPrefixKey+"\\", "", 1)
}
*/

// getProposal find proposal using id
func getProposal(id string) (*Proposal, error) {
	if val, ok := SystemProposal[id]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("proposal %s is not found", id)
}

func setProposal(proposal *Proposal) {
	SystemProposal[proposal.ID] = proposal
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
