/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package pmap

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/libp2p/go-libp2p-protocol"
)

const (
	// port for RPC
	DefaultRPCPort = 8915
	// port for query and register aergosvr
	DefaultSrvPort = 8916
)

// subprotocol for polaris
const (
	PolarisMapSub  protocol.ID = "/polaris/0.1"
	PolarisPingSub protocol.ID = "/ping/0.1"
)
const (
	MapQuery p2pcommon.SubProtocol = 0x0100 + iota
	MapResponse
)
