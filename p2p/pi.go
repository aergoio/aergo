/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"net"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
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
	ma, err := types.ToMultiAddr(ipAddress.String(), uint32(protocolPort))
	var meta p2pcommon.PeerMeta

	meta.ID = peerID
	meta.Role = setupPeerRole(produceBlock, conf)
	meta.Addresses = []types.Multiaddr{ma}
	switch meta.Role {
	case types.PeerRole_Producer:
		// register self id
		meta.ProducerIDs = []types.PeerID{peerID}
	case types.PeerRole_Agent:
		size := len(conf.Producers)
		if size == 0 {
			panic("invalid configuration: agent peer must have at least one producerID ")
		}
		pids := make([]types.PeerID, len(conf.Producers))
		for i, str := range conf.Producers {
			pid, err := types.IDB58Decode(str)
			if err != nil {
				panic("invalid producerID " + str + " : " + err.Error())
			}
			pids[i] = pid
		}
		meta.ProducerIDs = pids
	}
	meta.Hidden = !conf.NPExposeSelf
	meta.Version = p2pkey.NodeVersion()

	return meta
}

func setupPeerRole(enableBp bool, cfg *config.P2PConfig) types.PeerRole {
	roleInCfg := strings.ToLower(cfg.PeerRole)
	if enableBp {
		if len(roleInCfg) > 0 && roleInCfg != "producer" {
			panic("config mismatch. consensus.enablebp is true but p2p.peerrole is not producer")
		}
		return types.PeerRole_Producer
	} else { // role is blank, watcher or agent only
		switch roleInCfg {
		case "agent":
			return types.PeerRole_Agent
		case "watcher", "":
			return types.PeerRole_Watcher
		case "producer":
			panic("config mismatch. consensus.enablebp is false but p2p.peerrole is producer")

		default:
			panic("invalid p2p.peerrole config")
		}
	}
}
