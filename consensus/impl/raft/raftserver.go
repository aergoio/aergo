// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	raftlib "github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
)

//noinspection ALL
var raftLogger raftlib.Logger

func init() {
	raftLogger = NewRaftLogger(logger)
}

// A key-value stream backed by raft
type raftServer struct {
	proposeC    <-chan string            // proposed messages (k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan *string             // entries committed to log (k,v)
	errorC      chan error               // errors from raft session

	id          uint64   // client ID for raft session
	peers       []string // raft peer URLs
	join        bool     // node is joining an existing cluster
	waldir      string   // path to WAL directory
	snapdir     string   // path to snapshot directory
	getSnapshot func() ([]byte, error)
	lastIndex   uint64 // index of log at start

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raftlib.Node
	raftStorage *raftlib.MemoryStorage
	wal         *wal.WAL

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64
	transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete

	leaderStatus LeaderStatus

	certFile string
	keyFile  string
}

type LeaderStatus struct {
	leader        uint64
	leaderChanged uint64
}

var defaultSnapCount uint64 = 10000
var snapshotCatchUpEntriesN uint64 = 10000

// newRaftServer initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func newRaftServer(id uint64, peers []string, join bool, waldir string, snapdir string,
	certFile string, keyFile string,
	getSnapshot func() ([]byte, error), proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) *raftServer {

	commitC := make(chan *string)
	errorC := make(chan error)

	rs := &raftServer{
		proposeC:    proposeC,
		confChangeC: confChangeC,
		commitC:     commitC,
		errorC:      errorC,
		id:          id,
		peers:       peers,
		join:        join,
		waldir:      waldir,
		snapdir:     snapdir,
		getSnapshot: getSnapshot,
		snapCount:   defaultSnapCount,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
		// rest of structure populated after WAL replay

		certFile: certFile,
		keyFile:  keyFile,
	}
	go rs.startRaft()

	return rs
}

func (rs *raftServer) startRaft() {
	if !fileutil.Exist(rs.snapdir) {
		if err := os.MkdirAll(rs.snapdir, 0750); err != nil {
			logger.Error().Err(err).Msg("cannot create dir for snapshot")
		}
	}
	rs.snapshotter = snap.New(rs.snapdir)
	rs.snapshotterReady <- rs.snapshotter

	oldwal := wal.Exist(rs.waldir)
	rs.wal = rs.replayWAL()

	rpeers := make([]raftlib.Peer, len(rs.peers))
	for i := range rpeers {
		rpeers[i] = raftlib.Peer{ID: uint64(i + 1)}
	}
	c := &raftlib.Config{
		ID:              uint64(rs.id),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         rs.raftStorage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
		Logger:          raftLogger,
	}

	if oldwal {
		rs.node = raftlib.RestartNode(c)
	} else {
		startPeers := rpeers
		if rs.join {
			startPeers = nil
		}
		rs.node = raftlib.StartNode(c, startPeers)
	}

	rs.transport = &rafthttp.Transport{
		ID:          types.ID(rs.id),
		ClusterID:   0x1000,
		Raft:        rs,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.FormatUint(rs.id, 10)),
		ErrorC:      make(chan error),
	}

	rs.transport.Start()
	for i := range rs.peers {
		if uint64(i+1) != rs.id {
			rs.transport.AddPeer(types.ID(i+1), []string{rs.peers[i]})
		}
	}

	go rs.serveRaft()
	go rs.serveChannels()
}

// stop closes http, closes all channels, and stops raft.
func (rs *raftServer) stop() {
	rs.stopHTTP()
	close(rs.commitC)
	close(rs.errorC)
	rs.node.Stop()
}

func (rs *raftServer) stopHTTP() {
	rs.transport.Stop()
	close(rs.httpstopc)
	<-rs.httpdonec
}

func (rs *raftServer) writeError(err error) {
	rs.stopHTTP()
	close(rs.commitC)
	rs.errorC <- err
	close(rs.errorC)
	rs.node.Stop()
}

func (rs *raftServer) serveChannels() {
	snapshot, err := rs.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rs.confState = snapshot.Metadata.ConfState
	rs.snapshotIndex = snapshot.Metadata.Index
	rs.appliedIndex = snapshot.Metadata.Index

	defer rs.wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// send proposals over raft
	go func() {
		var confChangeCount uint64 = 0

		for rs.proposeC != nil && rs.confChangeC != nil {
			select {
			case prop, ok := <-rs.proposeC:
				if !ok {
					rs.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					rs.node.Propose(context.TODO(), []byte(prop))
				}

			case cc, ok := <-rs.confChangeC:
				if !ok {
					rs.confChangeC = nil
				} else {
					confChangeCount += 1
					cc.ID = confChangeCount
					rs.node.ProposeConfChange(context.TODO(), cc)
				}
			}
		}
		// client closed channel; shutdown raft if not already
		close(rs.stopc)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rs.node.Tick()

			// store raft entries to wal, then publish over commit channel
		case rd := <-rs.node.Ready():
			if rd.SoftState != nil {
				rs.updateLeader(rd.SoftState)
			}

			rs.wal.Save(rd.HardState, rd.Entries)
			if !raftlib.IsEmptySnap(rd.Snapshot) {
				panic("snapshot occurred!!")
				rs.saveSnap(rd.Snapshot)
				rs.raftStorage.ApplySnapshot(rd.Snapshot)
				rs.publishSnapshot(rd.Snapshot)
			}
			rs.raftStorage.Append(rd.Entries)
			rs.transport.Send(rd.Messages)
			if ok := rs.publishEntries(rs.entriesToApply(rd.CommittedEntries)); !ok {
				rs.stop()
				return
			}
			rs.maybeTriggerSnapshot()
			rs.node.Advance()

		case err := <-rs.transport.ErrorC:
			rs.writeError(err)
			return

		case <-rs.stopc:
			rs.stop()
			return
		}
	}
}

func (rs *raftServer) serveRaft() {
	urlstr := rs.peers[rs.id-1]
	urlData, err := url.Parse(urlstr)
	if err != nil {
		logger.Fatal().Err(err).Str("url", urlstr).Msg("Failed parsing URL")
	}

	ln, err := newStoppableListener(urlData.Host, rs.httpstopc)
	if err != nil {
		logger.Fatal().Err(err).Str("url", urlstr).Msg("Failed to listen rafthttp")
	}

	if len(rs.certFile) != 0 && len(rs.keyFile) != 0 {
		logger.Info().Str("url", urlstr).Str("certfile", rs.certFile).Str("keyfile", rs.keyFile).
			Msg("raft http server(tls) started")

		err = (&http.Server{Handler: rs.transport.Handler()}).ServeTLS(ln, rs.certFile, rs.keyFile)
	} else {
		logger.Info().Str("url", urlstr).Msg("raft http server started")

		err = (&http.Server{Handler: rs.transport.Handler()}).Serve(ln)
	}

	select {
	case <-rs.httpstopc:
	default:
		logger.Fatal().Err(err).Msg("Failed to serve rafthttp")
	}
	close(rs.httpdonec)
}

func (rs *raftServer) loadSnapshot() *raftpb.Snapshot {
	snapshot, err := rs.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		logger.Fatal().Err(err).Msg("error loading snapshot")
	}
	return snapshot
}

// openWAL returns a WAL ready for reading.
func (rs *raftServer) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rs.waldir) {
		if err := os.MkdirAll(rs.waldir, 0750); err != nil {
			logger.Fatal().Err(err).Msg("cannot create dir for wal")
		}

		w, err := wal.Create(rs.waldir, nil)
		if err != nil {
			logger.Fatal().Err(err).Msg("create wal error")
		}

		logger.Info().Str("dir", rs.waldir).Msg("create wal directory")
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	logger.Info().Uint64("term", walsnap.Term).Uint64("index", walsnap.Index).Msg("loading WAL at term %d and index")
	w, err := wal.Open(rs.waldir, walsnap)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading wal")
	}

	logger.Info().Msg("openwal done")
	return w
}

// replayWAL replays WAL entries into the raft instance.
func (rs *raftServer) replayWAL() *wal.WAL {
	logger.Info().Uint64("raftid", rs.id).Msg("replaying WAL of member %d")
	snapshot := rs.loadSnapshot()
	w := rs.openWAL(snapshot)
	_, st, ents, err := w.ReadAll()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
	}
	rs.raftStorage = raftlib.NewMemoryStorage()
	if snapshot != nil {
		rs.raftStorage.ApplySnapshot(*snapshot)
	}
	rs.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	rs.raftStorage.Append(ents)
	// send nil once lastIndex is published so client knows commit channel is current
	if len(ents) > 0 {
		rs.lastIndex = ents[len(ents)-1].Index
	} else {
		//commitChannel used for syncing startup
		rs.commitC <- nil
	}

	logger.Info().Msg("replaying WAL done")

	return w
}

func (rs *raftServer) maybeTriggerSnapshot() {
	if rs.appliedIndex-rs.snapshotIndex <= rs.snapCount {
		return
	}

	logger.Info().Uint64("applied index", rs.appliedIndex).Uint64("last snapshot index", rs.snapshotIndex).Msg("start snapshot")
	data, err := rs.getSnapshot()
	if err != nil {
		logger.Panic().Err(err).Msg("raft getsnapshot failed")
	}
	snapshot, err := rs.raftStorage.CreateSnapshot(rs.appliedIndex, &rs.confState, data)
	if err != nil {
		panic(err)
	}
	if err := rs.saveSnap(snapshot); err != nil {
		panic(err)
	}

	compactIndex := uint64(1)
	if rs.appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = rs.appliedIndex - snapshotCatchUpEntriesN
	}
	if err := rs.raftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}

	logger.Info().Uint64("index", compactIndex).Msg("compacted raftLog.at index")
	rs.snapshotIndex = rs.appliedIndex
}

func (rs *raftServer) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raftlib.IsEmptySnap(snapshotToSave) {
		return
	}

	logger.Info().Uint64("index", rs.snapshotIndex).Msg("publishing snapshot at index")
	defer logger.Info().Uint64("index", rs.snapshotIndex).Msg("finished publishing snapshot at index")

	if snapshotToSave.Metadata.Index <= rs.appliedIndex {
		logger.Fatal().Msgf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index, rs.appliedIndex)
	}
	rs.commitC <- nil // trigger kvstore to load snapshot

	rs.confState = snapshotToSave.Metadata.ConfState
	rs.snapshotIndex = snapshotToSave.Metadata.Index
	rs.appliedIndex = snapshotToSave.Metadata.Index
}

func (rs *raftServer) saveSnap(snap raftpb.Snapshot) error {
	// must save the snapshot index to the WAL before saving the
	// snapshot to maintain the invariant that we only Open the
	// wal at previously-saved snapshot indexes.
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := rs.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := rs.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	return rs.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (rs *raftServer) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return
	}
	firstIdx := ents[0].Index
	if firstIdx > rs.appliedIndex+1 {
		logger.Fatal().Msgf("first index of committed entry[%d] should <= progress.appliedIndex[%d] 1", firstIdx, rs.appliedIndex)
	}
	if rs.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rs.appliedIndex-firstIdx+1:]
	}
	return nents
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rs *raftServer) publishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}
			s := string(ents[i].Data)
			select {
			case rs.commitC <- &s:
			case <-rs.stopc:
				return false
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)
			rs.confState = *rs.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					rs.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rs.id) {
					logger.Info().Msg("I've been removed from the cluster! Shutting down.")
					return false
				}
				rs.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit, update appliedIndex
		rs.appliedIndex = ents[i].Index

		// special nil commit to signal replay has finished
		if ents[i].Index == rs.lastIndex {
			select {
			case rs.commitC <- nil:
			case <-rs.stopc:
				return false
			}
		}
	}
	return true
}

func (rs *raftServer) Process(ctx context.Context, m raftpb.Message) error {
	return rs.node.Step(ctx, m)
}
func (rs *raftServer) IsIDRemoved(id uint64) bool                              { return false }
func (rs *raftServer) ReportUnreachable(id uint64)                             {}
func (rs *raftServer) ReportSnapshot(id uint64, status raftlib.SnapshotStatus) {}

func (rs *raftServer) WaitStartup() {
	logger.Debug().Msg("raft start wait")
	for s := range rs.commitC {
		if s == nil {
			break
		}
	}
	logger.Debug().Msg("raft start succeed")
}

func (rs *raftServer) updateLeader(softState *raftlib.SoftState) {
	if softState.Lead != rs.GetLeader() {
		atomic.StoreUint64(&rs.leaderStatus.leader, softState.Lead)

		rs.leaderStatus.leaderChanged++

		logger.Info().Uint64("ID", rs.id).Uint64("leader", softState.Lead).Msg("leader changed")
	}
}

func (rs *raftServer) GetLeader() uint64 {
	return atomic.LoadUint64(&rs.leaderStatus.leader)
}

func (rs *raftServer) IsLeader() bool {
	return rs.id == rs.GetLeader()
}
