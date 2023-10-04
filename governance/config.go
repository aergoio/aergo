package governance

import (
	"github.com/aergoio/aergo/v2/governance/system"
	"github.com/aergoio/aergo/v2/types"
)

type Config struct {
	genesis         *types.Genesis
	consensusType   string
	cmds            map[types.OpSysTx]system.SysCmdCtor
	proposalCatalog *system.ProposalCatalog
	votingCatalog   []types.VotingIssue
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
	si.proposalCatalog = system.NewProposalCatalog(map[string]*system.Proposal{
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

	si.votingCatalog = make([]types.VotingIssue, 0, 10)
	si.votingCatalog = append(si.votingCatalog, types.GetVotingIssues()...)
	si.votingCatalog = append(si.votingCatalog, system.GetVotingIssues()...)
	return si
}
