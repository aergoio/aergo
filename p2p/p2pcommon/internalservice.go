/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/types"
	"net"
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
	//GetConsensusAccessor() consensus.ConsensusAccessor

	CertificateManager() CertificateManager

	RoleManager() PeerRoleManager
	// TODO add other methods
}

//go:generate mockgen -source=internalservice.go  -package=p2pmock -destination=../p2pmock/mock_internalservice.go

type LocalSettings struct {
	AgentID       types.PeerID
	InternalZones []*net.IPNet
}