package slot

import (
	"time"

	"github.com/aergoio/aergo/consensus/impl/dpos/bp"
)

var (
	// blockIntervalMs is the block genration interval in milli-seconds.
	blockIntervalMs int64
	// bpMinTimeLimitMs is the minimum block generation time limit in milli-sconds.
	bpMinTimeLimitMs int64
	// bpMaxTimeLimitMs is the maximum block generation time limit in milli-seconds.
	bpMaxTimeLimitMs int64
)

// Slot is a DPoS slot implmentation.
type Slot struct {
	timeNs    int64 // nanosecond
	timeMs    int64 // millisecond
	prevIndex int64
	nextIndex int64
}

// Init initilizes various slot parameters
func Init(blockIntervalSec int64) {
	blockIntervalMs = blockIntervalSec * 1000
	bpMinTimeLimitMs = blockIntervalMs / 4
	bpMaxTimeLimitMs = blockIntervalMs / 2
}

// Now returns a Slot corresponding to the current local time.
func Now() *Slot {
	return Time(time.Now())
}

// NewFromUnixNano returns a Slot corresponding to a UNIX time value (ns).
func NewFromUnixNano(ns int64) *Slot {
	return fromUnixNs(ns)
}

// UnixNano returns UNIX time in ns.
func (s *Slot) UnixNano() int64 {
	return s.timeNs
}

// Time returns a Slot corresponting to the given time.
func Time(t time.Time) *Slot {
	return fromUnixNs(t.UnixNano())
}

func fromUnixNs(ns int64) *Slot {
	ms := nsToMs(ns)
	return &Slot{
		timeNs:    ns,
		timeMs:    ms,
		prevIndex: msToPrevIndex(ms),
		nextIndex: msToNextIndex(ms),
	}
}

// IsValidNow reports whether s is still valid at the time when it's
// called.
func (s *Slot) IsValidNow() bool {
	if s.nextIndex == Now().nextIndex {
		return true
	}
	return false
}

// IsFuture reports whether s
func (s *Slot) IsFuture() bool {
	// Assume that the slot is future if the index difference >= 2.
	if s.nextIndex >= Now().nextIndex+2 {
		return true
	}
	return false
}

// Equal reports whether s1 has the same index as s2 or not.
func Equal(s1, s2 *Slot) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return s1.nextIndex == s2.nextIndex
}

// LessEqual reports whehter s1.nextIndex is less than or equal to s2.nextIndex
func LessEqual(s1, s2 *Slot) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return s1.nextIndex <= s2.nextIndex
}

// IsNextTo reports whether s1 corrensponds to the slot next to s2.
func IsNextTo(s1, s2 *Slot) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return s1.prevIndex == s2.nextIndex
}

// IsFor reports whether s correponds to myBpIdx (block producer index).
func (s *Slot) IsFor(bpIdx bp.Index, bpCount uint16) bool {
	return s.NextBpIndex(bpCount) == int64(bpIdx)
}

// GetBpTimeout returns the time available for block production.
func (s *Slot) GetBpTimeout() int64 {
	rTime := s.RemainingTimeMS()

	if rTime >= bpMaxTimeLimitMs {
		return bpMaxTimeLimitMs
	}

	return rTime
}

// RemainingTimeMS returns the remaining duration until the next block
// generation time.
func (s *Slot) RemainingTimeMS() int64 {
	return s.nextIndex*blockIntervalMs - nsToMs(time.Now().UnixNano())
}

// TimesUp reports whether the reminaing time <= BpMinTimeLimitMs
func (s Slot) TimesUp() bool {
	return s.RemainingTimeMS() <= bpMinTimeLimitMs
}

// NextBpIndex returns BP index for s.nextIndex.
func (s *Slot) NextBpIndex(bpCount uint16) int64 {
	return s.nextIndex % int64(bpCount)
}

func msToPrevIndex(ms int64) int64 {
	return msToIndex(ms)
}

func msToNextIndex(ms int64) int64 {
	return msToIndex(ms + blockIntervalMs)
}

func msToIndex(ms int64) int64 {
	return (ms - 1) / blockIntervalMs
}

func nsToMs(ns int64) int64 {
	return ns / 1000000
}

func msToSec(ms int64) int64 {
	return ms / 1000
}
