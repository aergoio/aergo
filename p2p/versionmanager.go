/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	v030 "github.com/aergoio/aergo/p2p/v030"
	"github.com/aergoio/aergo/types"
	peer2 "github.com/libp2p/go-libp2p-peer"
	"io"
)

type defaultVersionManager struct {
	pm     p2pcommon.PeerManager
	actor  p2pcommon.ActorService
	logger *log.Logger

	// check if is it adhoc
	localChainID *types.ChainID
}

func newDefaultVersionManager(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, logger *log.Logger, localChainID *types.ChainID) *defaultVersionManager {
	return &defaultVersionManager{pm: pm, actor: actor, logger: logger, localChainID: localChainID}
}

func (vm *defaultVersionManager) FindBestP2PVersion(versions []p2pcommon.P2PVersion) p2pcommon.P2PVersion {
	for _, suppored := range CurrentSupported {
		for _, reqVer := range versions {
			if suppored == reqVer {
				return reqVer
			}
		}
	}
	return p2pcommon.P2PVersionUnknown
}

func (h *defaultVersionManager) GetVersionedHandshaker(version p2pcommon.P2PVersion, peerID peer2.ID, r io.Reader, w io.Writer) (p2pcommon.VersionedHandshaker, error) {
	switch version {
	case p2pcommon.P2PVersion031:
		// TODO:
		v030hs := v030.NewV030StateHS(h.pm, h.actor, h.logger, h.localChainID, peerID, r, w)
		return v030hs, nil
	case p2pcommon.P2PVersion030:
		v030hs := v030.NewV030StateHS(h.pm, h.actor, h.logger, h.localChainID, peerID, r, w)
		return v030hs, nil
	default:
		return nil, fmt.Errorf("not supported version")
	}
}

func (vm *defaultVersionManager) InjectHandlers(version p2pcommon.P2PVersion, peer p2pcommon.RemotePeer) {
	panic("implement me")
}

