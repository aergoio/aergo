/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package factory

import (
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/impl/dpos"
	"github.com/aergoio/aergo/consensus/impl/sbp"
	"github.com/aergoio/aergo/pkg/component"
)

// New returns consensus.Consensus based on the configuration parameters.
func New(cfg *config.Config, hub *component.ComponentHub) (consensus.Consensus, error) {
	var c consensus.Consensus
	var err error

	if cfg.Consensus.EnableDpos {
		c, err = dpos.New(cfg, hub)
	} else {
		c, err = sbp.New(cfg, hub)
	}

	return c, err
}
