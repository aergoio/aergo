/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
)

var (
	UseFullID bool
)

// ShortForm returns shorthanded types.PeerID.
func ShortForm(pid types.PeerID) string {
	pretty := pid.Pretty()
	if UseFullID {
		return pretty
	}
	if len(pretty) > 10 {
		return fmt.Sprintf("%s*%s", pretty[:2], pretty[len(pretty)-6:])
	} else {
		return pretty
	}

}

func ShortMetaForm(m p2pcommon.PeerMeta) string {
	return fmt.Sprintf("%s:%d/%s", m.PrimaryAddress(), m.PrimaryPort(), ShortForm(m.ID))
}

func ContainsID(pool []types.PeerID, id types.PeerID) bool {
	for _, pid := range pool {
		if types.IsSamePeerID(pid, id) {
			return true
		}
	}
	return false
}
