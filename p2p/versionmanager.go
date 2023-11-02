/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"io"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	v030 "github.com/aergoio/aergo/v2/p2p/v030"
	v200 "github.com/aergoio/aergo/v2/p2p/v200"
	"github.com/aergoio/aergo/v2/types"
)

type defaultVersionManager struct {
	is     p2pcommon.InternalService
	pm     p2pcommon.PeerManager
	actor  p2pcommon.ActorService
	ca     types.ChainAccessor
	logger *log.Logger

	// check if is it ad hoc
	localChainID *types.ChainID
}

func newDefaultVersionManager(is p2pcommon.InternalService, actor p2pcommon.ActorService, pm p2pcommon.PeerManager, ca types.ChainAccessor, logger *log.Logger, localChainID *types.ChainID) *defaultVersionManager {
	return &defaultVersionManager{is: is, pm: pm, actor: actor, ca: ca, logger: logger, localChainID: localChainID}
}

func (vm *defaultVersionManager) FindBestP2PVersion(versions []p2pcommon.P2PVersion) p2pcommon.P2PVersion {
	for _, supported := range p2pcommon.AcceptedInboundVersions {
		for _, reqVer := range versions {
			if supported == reqVer {
				return reqVer
			}
		}
	}
	return p2pcommon.P2PVersionUnknown
}

func (vm *defaultVersionManager) GetVersionedHandshaker(version p2pcommon.P2PVersion, peerID types.PeerID, rwc io.ReadWriteCloser) (p2pcommon.VersionedHandshaker, error) {
	switch version {
	case p2pcommon.P2PVersion200:
		vhs := v200.NewV200VersionedHS(vm.is, vm.logger, vm, vm.is.CertificateManager(), peerID, rwc, chain.Genesis.Block().Hash)
		return vhs, nil
	case p2pcommon.P2PVersion033:
		vhs := v030.NewV033VersionedHS(vm.pm, vm.actor, vm.logger, vm, peerID, rwc, chain.Genesis.Block().Hash)
		return vhs, nil
	case p2pcommon.P2PVersion032:
		vhs := v030.NewV032VersionedHS(vm.pm, vm.actor, vm.logger, vm.localChainID, peerID, rwc, chain.Genesis.Block().Hash)
		return vhs, nil
	case p2pcommon.P2PVersion031:
		v030hs := v030.NewV030VersionedHS(vm.pm, vm.actor, vm.logger, vm.localChainID, peerID, rwc)
		return v030hs, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (vm *defaultVersionManager) GetBestChainID() *types.ChainID {
	bb, _ := vm.ca.GetBestBlock() // error is always nil at current version
	if bb != nil {
		return vm.ca.ChainID(bb.BlockNo())
	} else {
		return nil
	}
}

func (vm *defaultVersionManager) GetChainID(no types.BlockNo) *types.ChainID {
	return vm.ca.ChainID(no)
}
