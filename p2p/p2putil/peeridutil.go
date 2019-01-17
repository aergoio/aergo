/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"github.com/libp2p/go-libp2p-peer"
)

// ShortForm returns shorthanded peer.ID.
func ShortForm(pid peer.ID) string {
	pretty := pid.Pretty()
	if len(pretty) > 10 {
		return fmt.Sprintf("%s..%s", pretty[:2], pretty[len(pretty)-6:])
	} else {
		return pretty
	}

}
