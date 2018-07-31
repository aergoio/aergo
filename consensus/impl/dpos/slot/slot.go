package slot

import (
	"time"

	"github.com/aergoio/aergo/consensus/impl/dpos/param"
)

// Slot is a DPoS slot implmentation
type Slot struct {
	timeSec   int64
	timeMs    int64
	prevIndex int64
	nextIndex int64
}

// Now returns a Slot corresponding to the current local time.
func Now() *Slot {
	return Time(time.Now())
}

// Unix returns a Slot corresponding to a UNIX time value (s).
func Unix(sec int64) *Slot {
	return fromUnixMs(sec * 1000)
}

// Time returns a Slot corresponting to the given time.
func Time(t time.Time) *Slot {
	return fromUnixMs(nsToMs(t.UnixNano()))
}

func fromUnixMs(ms int64) *Slot {
	return &Slot{
		timeSec:   msToSec(ms),
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

// GetBpTimeout returns the time avaliable for block production.
func (s *Slot) GetBpTimeout() int64 {
	rTime := s.RemainingTimeMS()

	if rTime >= param.BpMaxTimeLimitMs {
		return param.BpMaxTimeLimitMs
	}

	return rTime
}

// RemainingTimeMS returns the remaining duration until the next block
// generation time.
func (s *Slot) RemainingTimeMS() int64 {
	return s.nextIndex*param.BlockIntervalMs - nsToMs(time.Now().UnixNano())
}

func (s *Slot) nextBpIndex() int64 {
	return absToBpIndex(s.nextIndex)
}

func absToBpIndex(idx int64) int64 {
	return idx % param.BlockProducers
}

func msToPrevIndex(ms int64) int64 {
	return msToIndex(ms)
}

func msToNextIndex(ms int64) int64 {
	return msToIndex(ms + param.BlockIntervalMs)
}

func msToIndex(ms int64) int64 {
	return (ms - 1) / param.BlockIntervalMs
}

func nsToMs(ns int64) int64 {
	return ns / 1000000
}

func msToSec(ms int64) int64 {
	return ms / 1000
}
