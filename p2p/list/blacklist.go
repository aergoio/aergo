/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package list

import (
	"time"
)

// BanStatus keep kickout logs and decide how long the ban duration is
type BanStatus interface {
	// ID is ip address or peer id
	ID() string

	// BanUntil show when this ban items is expired.
	BanUntil() time.Time
	Banned(refTime time.Time) bool
	Events() []BanEvent
	PruneOldEvents(pruneTime time.Time) int
}

type BanEvent interface {
	When() time.Time
	Why() string
}

// BanThreshold is number of events to ban address or peerid
const BanThreshold = 5

var BanDurations = []time.Duration{
	0,
	0,
	time.Minute,
	time.Minute * 3,
	time.Minute * 10,
	time.Hour,
	time.Hour * 24,
	time.Hour * 24 * 30,
	time.Hour * 24 * 3650,
}

const BanValidDuration = time.Minute * 30
const BanReleaseDuration = time.Hour * 24 * 730
