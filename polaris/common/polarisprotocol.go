/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package common

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	core "github.com/libp2p/go-libp2p-core"
	"time"
)

const (
	// port for RPC
	DefaultRPCPort = 8915
	// port for query and register aergosvr
	DefaultSrvPort = 8916
)

// subprotocol for polaris
const (
	PolarisMapSub  core.ProtocolID = "/polaris/0.1"
	PolarisPingSub core.ProtocolID = "/ping/0.1"
)
const (
	MapQuery p2pcommon.SubProtocol = 0x0100 + iota
	MapResponse
)

const PolarisConnectionTTL = time.Second * 30

