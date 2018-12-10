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
func New(cfg *config.Config, cs *chain.ChainService, hub *component.ComponentHub) (consensus.Consensus, error) {
	var c consensus.Consensus
	var err error

	if cfg.Consensus.EnableDpos {
		c, err = dpos.New(cfg, cs.CDBReader(), hub)
	} else {
		c, err = sbp.New(cfg, hub)
	}

	if err == nil {
		// Link mutual references.
		cs.SetChainConsensus(c)
		c.SetStateDB(cs.SDB())
		c.SetChainAccessor(cs)
	}

	return c, err
}
