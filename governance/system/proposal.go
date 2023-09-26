package system

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

const (
	proposalPrefixKey = "proposal" //aergo proposal format
)

type Proposals struct {
	proposal map[string]*Proposal
}

func NewProposals(init map[string]*Proposal) *Proposals {
	p := &Proposals{
		proposal: make(map[string]*Proposal),
	}
	if init != nil {
		p.proposal = init
	}
	return p
}

func (p *Proposals) GetProposal(id string) (*Proposal, error) {
	if proposal, ok := p.proposal[strings.ToUpper(id)]; ok {
		return proposal, nil
	}
	return nil, fmt.Errorf("proposal %s is not found", id)
}

func (p *Proposals) SetProposal(proposal *Proposal) *Proposal {
	p.proposal[strings.ToUpper(proposal.ID)] = proposal
	return proposal
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
	for i := SysParamIndex(0); i < SysParamMax; i++ {
		if strings.ToUpper(id) == i.ID() {
			return true
		}
	}
	return false
}
