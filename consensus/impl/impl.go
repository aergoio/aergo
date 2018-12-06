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
		cdb  = cs.CDBReader()
		impl = map[string]consensus.Constructor{
			"dpos": dpos.GetConstructor(cfg, hub, cdb), // DPoS
			"sbp":  sbp.GetConstructor(cfg, hub, cdb),  // Simple BP
		}

		c   consensus.Consensus
		err error
	)

	if cfg.Consensus.EnableDpos {
		c, err = impl["dpos"]()
	} else {
		c, err = impl["sbp"]()
	}

	if err == nil {
		// Link mutual references.
		cs.SetChainConsensus(c)
		c.SetStateDB(cs.SDB())
		c.SetChainAccessor(cs)
	}

	return c, err
}
