package governance

import (
	"github.com/aergoio/aergo/v2/governance/system"
	"github.com/aergoio/aergo/v2/types"
)

type Config struct {
	genesis        *types.Genesis
	consensusType  string
	cmds           map[types.OpSysTx]system.SysCmdCtor
	proposals      *system.Proposals
	votingCatalogs []types.VotingIssue
}

func NewConfig(genesis *types.Genesis, consensus string) *Config {
	si := &Config{}
	si.genesis = genesis
	si.consensusType = consensus
	si.cmds = map[types.OpSysTx]system.SysCmdCtor{
		types.OpvoteBP:  system.NewVoteCmd,
		types.OpvoteDAO: system.NewVoteCmd,
		types.Opstake:   system.NewStakeCmd,
		types.Opunstake: system.NewUnstakeCmd,
	}
	si.proposals = system.NewProposals(map[string]*system.Proposal{
		system.BpCount.ID(): {
			ID:             system.BpCount.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		system.StakingMin.ID(): {
			ID:             system.StakingMin.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		system.GasPrice.ID(): {
			ID:             system.GasPrice.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		system.NamePrice.ID(): {
			ID:             system.NamePrice.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
	})

	si.votingCatalogs = make([]types.VotingIssue, 0, 10)
	si.votingCatalogs = append(si.votingCatalogs, types.GetVotingIssues()...)
	si.votingCatalogs = append(si.votingCatalogs, system.GetVotingIssues()...)
	return si
}

// getProposal find proposal using id
func (c *Config) GetProposal(id string) (*system.Proposal, error) {
	return c.proposals.GetProposal(id)
}

func (c *Config) SetProposal(proposal *system.Proposal) {
	c.proposals.SetProposal(proposal)
}
