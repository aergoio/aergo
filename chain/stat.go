package chain

import (
	"sync"
)

const (
	// TODO: Generate code converting the constants below to the corresponding
	// string by using 'go generate' command.

	// StatReorg is a constant representing a stat about reorganization.
	statReorg = iota
	// StatMax is a constant representing a value less than which all the
	// constants corresponding chain stats must be.
	statMax
)

var (
	// To add a new one to chain stats, implements statItem interface and add
	// its constructor here. Additionally you need to add a constant
	// corresponding to its index like statReorg above.
	statItemCtors = map[int]func() statItem{
		statReorg: newStReorg,
	}
)

type stats []*stat

func newStats() stats {
	s := make(stats, statMax)
	for i := 0; i < statMax; i++ {
		s[i] = newStat(statItemCtors[i]())
	}
	return s
}

func (s stats) get(statIdx int) *stat {
	return []*stat(s)[statIdx]
}

func (s stats) clone(statIdx int) interface{} {
	i := s.get(statIdx)
	i.RLock()
	defer i.RUnlock()
	return i.clone()
}

func (s stats) getCount(statIdx int) int64 {
	i := s.get(statIdx)
	i.RLock()
	defer i.RUnlock()

	return i.getCount()
}

func (s stats) getLastestEvent(statIdx int) interface{} {
	i := s.get(statIdx)
	i.RLock()
	defer i.RUnlock()

	return i.getLatestEvent()
}

func (s stats) updateEvent(statIdx int, args ...interface{}) {
	i := s.get(statIdx)
	i.Lock()
	defer i.Unlock()

	i.updateEvent(args...)
}

type stat struct {
	sync.RWMutex
	statItem
}

func newStat(i statItem) *stat {
	return &stat{statItem: i}
}

type statItem interface {
	getCount() int64
	getLatestEvent() interface{}
	updateEvent(args ...interface{})
	clone() interface{}
}

type stReorg struct {
	Count  int64
	Latest *evReorg
}

func newStReorg() statItem {
	return &stReorg{}
}

type evReorg struct {
	OldBest string
	NewBest string
	Branch  string
}

func (sr *stReorg) getCount() int64 {
	return sr.Count
}

func (sr *stReorg) getLatestEvent() interface{} {
	return sr.Latest
}

func (sr *stReorg) updateEvent(args ...interface{}) {
	if len(args) != 3 {
		logger.Debug().Int("len", len(args)).Msg("invalid arguments for the reorg stat update")
		return
	}

	s := make([]string, len(args))
	for i, a := range args {
		ok := false
		if s[i], ok = a.(string); !ok {
			logger.Debug().Int("arg idx", i).Msg("non-string arguemt for the reorg stat update")
			return
		}
	}

	sr.Latest = &evReorg{
		OldBest: s[0],
		NewBest: s[1],
		Branch:  s[2],
	}
	sr.Count++
}

func (sr *stReorg) clone() interface{} {
	c := *sr
	if sr.Latest != nil {
		l := *sr.Latest
		c.Latest = &l
	}

	return &c
}
