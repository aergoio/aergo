/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"time"
)

// HostAccessor is interface to provide the informations about server
type HostAccessor interface {
	// Version return version of this server
	Version() string

	// StartTime is the time when server was booted
	StartTime() time.Time
}

