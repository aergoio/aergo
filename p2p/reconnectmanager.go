package p2p

import (
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"sync"

	"github.com/aergoio/aergo-lib/log"
	"github.com/libp2p/go-libp2p-peer"
)

// ReconnectManager manage reconnect job schedule
type ReconnectManager interface {
	AddJob(meta p2pcommon.PeerMeta)
	// CancelJob cancel from outer module to reconnectRunner
	CancelJob(pid peer.ID)
	// jobFinished remove reconnectRunner, which finish job for itself.
	jobFinished(pid peer.ID)

	Stop()
}

type reconnectManager struct {
	stop   bool
	pm     PeerManager
	logger *log.Logger
	mutex  *sync.Mutex

	jobs map[peer.ID]*reconnectJob
}

// newReconnectManager create partial-inited manager for reconnect peer.
// Note: it returns incomplete object, caller should set peerManager before using this.
func newReconnectManager(logger *log.Logger) *reconnectManager {
	return &reconnectManager{mutex: &sync.Mutex{}, jobs: make(map[peer.ID]*reconnectJob), logger: logger}
}

// AddJob add jobber to reconnect peer.
func (rm *reconnectManager) AddJob(meta p2pcommon.PeerMeta) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	if _, exist := rm.jobs[meta.ID]; exist || rm.stop {
		return
	}
	rm.logger.Debug().Str(LogPeerID, p2putil.ShortForm(meta.ID)).Msg("Starting reconnect job")
	jobRunner := newReconnectRunner(meta, rm, rm.pm, rm.logger)
	go jobRunner.runJob()
	rm.jobs[meta.ID] = jobRunner
}

// CancelJob cancel currently running AddJob
func (rm *reconnectManager) CancelJob(pid peer.ID) {
	rm.mutex.Lock()
	job, exist := rm.jobs[pid]
	if !exist {
		rm.mutex.Unlock()
		return
	}
	rm.logger.Debug().Str(LogPeerID, p2putil.ShortForm(pid)).Msg("Canceling reconnect job")
	delete(rm.jobs, pid)
	rm.mutex.Unlock()
	job.cancel <- struct{}{}
}

func (rm *reconnectManager) Stop() {
	rm.mutex.Lock()
	keys := make([]peer.ID, len(rm.jobs))
	i := 0
	for k := range rm.jobs {
		keys[i] = k
		i++
	}
	rm.stop = true
	rm.mutex.Unlock()
	rm.logger.Debug().Msg("Stopping reconnect manager")
	for _, key := range keys {
		rm.CancelJob(key)
	}
}

func (rm *reconnectManager) jobFinished(pid peer.ID) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.logger.Debug().Str(LogPeerID, p2putil.ShortForm(pid)).Msg("Clearing finished reconnect job")
	delete(rm.jobs, pid)
}
