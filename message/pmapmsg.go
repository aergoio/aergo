/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package message

import (
	"github.com/aergoio/aergo/types"
)

type MapQueryMsg struct {
	Count int
	BestBlock *types.Block
}

type MapQueryRsp struct {
	Peers []*types.PeerAddress
	Err error
}