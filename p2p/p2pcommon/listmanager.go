/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"time"

	"github.com/aergoio/aergo/v2/types"
)

// ListManager manages whitelist and blacklist
type ListManager interface {
	Start()
	Stop()

	IsBanned(addr string, pid types.PeerID) (bool, time.Time)

	// RefineList update white/blacklist
	RefineList()
	Summary() map[string]interface{}
}

//go:generate mockgen -source=listmanager.go -package=p2pmock -destination=../p2pmock/mock_listmanager.go
