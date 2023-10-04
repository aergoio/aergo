package governance

import (
	"math/big"

	"github.com/aergoio/aergo/v2/governance/system"
	"github.com/aergoio/aergo/v2/types"
)

type Config struct {
	genesis         *types.Genesis
	consensusType   string
	cmds            map[types.OpSysTx]system.SysCmdCtor
	proposalCatalog *system.ProposalCatalog
	votingCatalog   []types.VotingIssue
	params          *system.Parameters
}

func NewConfig(genesis *types.Genesis, consensus string) *Config {
	c := &Config{}
	c.genesis = genesis
	c.consensusType = consensus

	// init cmds
	c.cmds = map[types.OpSysTx]system.SysCmdCtor{
		types.OpvoteBP:  system.NewVoteCmd,
		types.OpvoteDAO: system.NewVoteCmd,
		types.Opstake:   system.NewStakeCmd,
		types.Opunstake: system.NewUnstakeCmd,
	}

	// init default params
	c.params = system.NewParameters(map[string]*big.Int{
		system.StakingMin.ID(): types.StakingMinimum,
		system.GasPrice.ID():   types.NewAmount(50, types.Gaer), // 50 gaer
		system.NamePrice.ID():  types.NewAmount(1, types.Aergo), // 1 aergo
	})

	// init proposal catalog
	c.proposalCatalog = system.NewProposalCatalog(map[string]*system.Proposal{
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

	// init voting catalog
	c.votingCatalog = make([]types.VotingIssue, 0, 10)
	c.votingCatalog = append(c.votingCatalog, types.GetVotingIssues()...)
	c.votingCatalog = append(c.votingCatalog, system.GetVotingIssues()...)

	return c
}
