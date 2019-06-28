/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"github.com/aergoio/aergo/types"
	"github.com/rs/zerolog"
)

type LogGetBlockRespMashaler struct {
	*types.GetBlockResponse
}

func (m LogGetBlockRespMashaler) MarshalZerologObject(e *zerolog.Event) {
	e.Str("status", m.Status.String()).Bool("hasNext", m.HasNext).Str("hashes", PrintHashList(m.Blocks))
}

