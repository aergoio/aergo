/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package impl

import (
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos"
	"github.com/aergoio/aergo/consensus/impl/raft"
	"github.com/aergoio/aergo/consensus/impl/sbp"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
)

// New returns consensus.Consensus based on the configuration parameters.
func New(cfg *config.Config, hub *component.ComponentHub, cs *chain.ChainService, pa types.PeerAccessor) (consensus.Consensus, error) {
	var (
		c   consensus.Consensus
		err error

		blockInterval int64
	)

	if chain.IsPublic() {
		blockInterval = 1
	} else {
		blockInterval = cfg.Consensus.BlockInterval
	}

	consensus.InitBlockInterval(blockInterval)

	if c, err = newConsensus(cfg, hub, cs); err == nil {
		// Link mutual references.
		cs.SetChainConsensus(c)
	}

	return c, err
}

func newConsensus(cfg *config.Config, hub *component.ComponentHub,
	cs *chain.ChainService) (consensus.Consensus, error) {
	cdb := cs.CDB()
	sdb := cs.SDB()

	impl := map[string]consensus.Constructor{
		dpos.GetName(): dpos.GetConstructor(cfg, hub, cdb, sdb), // DPoS
		sbp.GetName():  sbp.GetConstructor(cfg, hub, cdb, sdb),  // Simple BP
		raft.GetName(): raft.GetConstructor(cfg, hub, cdb, sdb), // Raft BP
	}

	return impl[cdb.GetGenesisInfo().ConsensusType()]()
}
