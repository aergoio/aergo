/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/internal/network"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"net"
	"strconv"
)

func SetupSelfMeta(peerID types.PeerID, conf *config.P2PConfig, produceBlock bool) p2pcommon.PeerMeta {
	protocolAddr := conf.NetProtocolAddr
	var ipAddress net.IP
	var err error
	var protocolPort int
	protocolPort = conf.NetProtocolPort
	if protocolPort <= 0 {
		panic("invalid NetProtocolPort " + strconv.Itoa(conf.NetProtocolPort))
	}
	if len(conf.NetProtocolAddr) != 0 {
		ipAddress, err = network.GetSingleIPAddress(protocolAddr)
		if err != nil {
			panic("Invalid protocol address " + protocolAddr + " : " + err.Error())
		}
		if ipAddress.IsUnspecified() {
			panic("NetProtocolAddr should be a specified IP address, not 0.0.0.0")
		}
	} else {
		extIP, err := p2putil.ExternalIP()
		if err != nil {
			panic("error while finding IP address: " + err.Error())
		}
		ipAddress = extIP
		protocolAddr = ipAddress.String()
	}
	ma,err := types.ToMultiAddr(ipAddress.String(), uint32(protocolPort))
	var meta p2pcommon.PeerMeta

	meta.ID = peerID
	meta.Addresses = []types.Multiaddr{ma}
	// TODO
	if produceBlock {
		meta.Role = types.PeerRole_Producer
		// register self id
		meta.ProducerIDs = []types.PeerID{peerID}
	} else {
		meta.Role = types.PeerRole_Watcher
	}
	meta.Hidden = !conf.NPExposeSelf
	meta.Version = p2pkey.NodeVersion()

	return meta
}
