/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
)

type MapPeerManager struct {
	PeerManager
}

func NewMapPeerManager(manager PeerManager) PeerManager {
	mapPM := &MapPeerManager{manager}

	return mapPM
}
