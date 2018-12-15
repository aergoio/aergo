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
	"github.com/aergoio/aergo/consensus/impl/sbp"
	"github.com/aergoio/aergo/pkg/component"
)

// New returns consensus.Consensus based on the configuration parameters.
func New(cfg *config.Config, hub *component.ComponentHub, cs *chain.ChainService) (consensus.Consensus, error) {
	var (
		c   consensus.Consensus
		err error
	)

	if c, err = newConsensus(cfg, hub, cs.CDBReader()); err == nil {
		// Link mutual references.
		cs.SetChainConsensus(c)
		c.SetStateDB(cs.SDB())
		c.SetChainAccessor(cs)
	}

	return c, err
}

func newConsensus(cfg *config.Config, hub *component.ComponentHub, cdb consensus.ChainDbReader) (consensus.Consensus, error) {
	impl := map[string]consensus.Constructor{
		"dpos": dpos.GetConstructor(cfg, hub, cdb), // DPoS
		"sbp":  sbp.GetConstructor(cfg, hub, cdb),  // Simple BP
	}

	return impl[cdb.GetGenesisInfo().Consensus()]()
}
