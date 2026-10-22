/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package impl

import (
	"strings"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/impl/dpos"
	"github.com/aergoio/aergo/v2/consensus/impl/raftv2"
	"github.com/aergoio/aergo/v2/consensus/impl/sbp"
	"github.com/aergoio/aergo/v2/p2p"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/rpc"
	"github.com/aergoio/aergo/v2/types"
)

// New returns consensus.Consensus based on the configuration parameters.
func New(cfg *config.Config, hub *component.ComponentHub, cs *chain.ChainService, p2psvc *p2p.P2P, rpcSvc *rpc.RPC) (consensus.Consensus, error) {
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

	if c, err = newConsensus(cfg, hub, cs, p2psvc.GetPeerAccessor()); err == nil {
		// Link mutual references.
		cs.SetChainConsensus(c)
		rpcSvc.SetConsensusAccessor(c)
		p2psvc.SetConsensusAccessor(c)
	}

	return c, err
}

func newConsensus(cfg *config.Config, hub *component.ComponentHub,
	cs *chain.ChainService, pa p2pcommon.PeerAccessor) (consensus.Consensus, error) {
	cdb := cs.CDB()
	sdb := cs.SDB()

	impl := map[string]consensus.Constructor{
		dpos.GetName():   dpos.GetConstructor(cfg, hub, cdb, sdb),              // DPoS
		sbp.GetName():    sbp.GetConstructor(cfg, hub, cdb, sdb),               // Simple BP
		raftv2.GetName(): raftv2.GetConstructor(cfg, hub, cs.WalDB(), sdb, pa), // Raft BP
	}

	consensus.SetCurConsensus(cdb.GetGenesisInfo().ConsensusType())
	return impl[cdb.GetGenesisInfo().ConsensusType()]()
}

type GenesisValidator func(genesis *types.Genesis) error

func ValidateGenesis(genesis *types.Genesis) error {
	name := strings.ToLower(genesis.ConsensusType())

	validators := map[string]GenesisValidator{
		dpos.GetName():   dpos.ValidateGenesis,   // DPoS
		sbp.GetName():    sbp.ValidateGenesis,    // Simple BP
		raftv2.GetName(): raftv2.ValidateGenesis, // Raft BP
	}

	return validators[name](genesis)
}
