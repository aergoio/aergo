package chain

import (
	"sync"
	"time"

	"github.com/aergoio/aergo/types"
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
	OldBest *blockInfo
	NewBest *blockInfo
	Branch  *blockInfo
	time    time.Time
}

type blockInfo struct {
	Hash   string
	Height types.BlockNo
}

func (sr *stReorg) getCount() int64 {
	return sr.Count
}

func (sr *stReorg) getLatestEvent() interface{} {
	return sr.Latest
}

func (sr *stReorg) updateEvent(args ...interface{}) {
	if len(args) != 3 {
		logger.Info().Int("len", len(args)).Msg("invalid # of arguments for the reorg stat update")
		return
	}

	bi := make([]*blockInfo, len(args))
	for i, a := range args {
		var block *types.Block
		ok := false
		if block, ok = a.(*types.Block); !ok {
			logger.Info().Int("arg idx", i).Msg("invalid type of argument")
			return
		}
		bi[i] = &blockInfo{Hash: block.ID(), Height: block.BlockNo()}
	}

	sr.Latest = &evReorg{
		OldBest: bi[0],
		NewBest: bi[1],
		Branch:  bi[2],
		time:    time.Now(),
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
