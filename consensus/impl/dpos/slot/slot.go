package slot

import (
	"time"

	"github.com/aergoio/aergo/consensus"
)

var (
	// BlockIntervalMs is the block genration interval in milli-seconds.
	BlockIntervalMs int64
	// BpMinTimeLimitMs is the minimum block generation time limit in milli-sconds.
	BpMinTimeLimitMs int64
	// BpMaxTimeLimitMs is the maximum block generation time limit in milli-seconds.
	BpMaxTimeLimitMs int64

	blockProducers uint16
)

// Slot is a DPoS slot implmentation.
type Slot struct {
	timeNs    int64 // nanosecond
	timeMs    int64 // millisecond
	prevIndex int64
	nextIndex int64
}

// Init initilizes various slot parameters
func Init(blockIntervalSec int64, bps uint16) {
	BlockIntervalMs = consensus.BlockIntervalSec * 1000
	BpMinTimeLimitMs = BlockIntervalMs / 4
	BpMaxTimeLimitMs = BlockIntervalMs / 2
	blockProducers = bps
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

// IsValidNow reports whether the Slot is still valid at the time when it's
// called.
func (s *Slot) IsValidNow() bool {
	if s.nextIndex == Now().nextIndex {
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
func (s *Slot) IsFor(bpIdx uint16) bool {
	return s.nextBpIndex() == int64(bpIdx)
}

// GetBpTimeout returns the time available for block production.
func (s *Slot) GetBpTimeout() int64 {
	rTime := s.RemainingTimeMS()

	if rTime >= BpMaxTimeLimitMs {
		return BpMaxTimeLimitMs
	}

	return rTime
}

// RemainingTimeMS returns the remaining duration until the next block
// generation time.
func (s *Slot) RemainingTimeMS() int64 {
	return s.nextIndex*BlockIntervalMs - nsToMs(time.Now().UnixNano())
}

// TimesUp reports whether the reminaing time <= BpMinTimeLimitMs
func (s Slot) TimesUp() bool {
	return s.RemainingTimeMS() <= BpMinTimeLimitMs
}

func (s *Slot) nextBpIndex() int64 {
	return absToBpIndex(s.nextIndex)
}

func absToBpIndex(idx int64) int64 {
	return idx % int64(blockProducers)
}

func msToPrevIndex(ms int64) int64 {
	return msToIndex(ms)
}

func msToNextIndex(ms int64) int64 {
	return msToIndex(ms + BlockIntervalMs)
}

func msToIndex(ms int64) int64 {
	return (ms - 1) / BlockIntervalMs
}

func nsToMs(ns int64) int64 {
	return ns / 1000000
}

func msToSec(ms int64) int64 {
	return ms / 1000
}
