/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"net"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/types"
)

// InternalService provides informations of self node and reference of other p2p components.
// This service is intended to be used inside p2p modules, and using outside of p2p is not expected.
type InternalService interface {
	//NetworkTransport
	SelfMeta() PeerMeta
	SelfNodeID() types.PeerID
	LocalSettings() LocalSettings

	// accessors of other modules
	GetChainAccessor() types.ChainAccessor
	ConsensusAccessor() consensus.ConsensusAccessor

	PeerManager() PeerManager

	CertificateManager() CertificateManager

	RoleManager() PeerRoleManager
}

//go:generate mockgen -source=internalservice.go  -package=p2pmock -destination=../p2pmock/mock_internalservice.go

type LocalSettings struct {
	AgentID       types.PeerID
	InternalZones []*net.IPNet
}
