/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/types"
)

const PolarisRPCSvc = "pRpcSvc"
const PolarisSvc = "polarisSvc"

type MapQueryMsg struct {
	Count int
	BestBlock *types.Block
}

type MapQueryRsp struct {
	Peers []*types.PeerAddress
	Err error
}

type PaginationMsg struct {
	ReferenceHash []byte
	Size          uint32
}


type CurrentListMsg PaginationMsg
type WhiteListMsg PaginationMsg
type BlackListMsg PaginationMsg
