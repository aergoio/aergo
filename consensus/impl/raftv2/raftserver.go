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

package raftv2

import (
	"bytes"
	"context"
	"errors"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-peer"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"

	"github.com/aergoio/etcd/etcdserver/stats"
	"github.com/aergoio/etcd/pkg/fileutil"
	etcdtypes "github.com/aergoio/etcd/pkg/types"
	raftlib "github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/rafthttp"
	"github.com/aergoio/etcd/snap"
)

//noinspection ALL
var (
	raftLogger              raftlib.Logger
	defaultSnapCount        uint64 = 10
	snapshotCatchUpEntriesN uint64 = 10
	ErrNoSnapshot                  = errors.New("no snapshot")
)

const (
	HasNoLeader uint64 = 0
)

func init() {
	raftLogger = NewRaftLogger(logger)
}

// A key-value stream backed by raft
type raftServer struct {
	*component.ComponentHub

	cluster *Cluster

	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan *types.Block        // entries committed to log (k,v)
	errorC      chan error               // errors from raft session

	id          uint64   // client ID for raft session
	peers       []string // raft peer URLs
	listenUrl   string
	join        bool   // node is joining an existing cluster
	snapdir     string // path to snapshot directory
	getSnapshot func() ([]byte, error)
	lastIndex   uint64 // index of log at start

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raftlib.Node
	raftStorage *raftlib.MemoryStorage
	//wal         *wal.WAL
	walDB *WalDB

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

	startSync  bool // maybe this flag is unnecessary
	lock       sync.RWMutex
	promotable bool

	tickMS time.Duration

	lastAppliedBlock *types.Block //tracking last applied block. It's initillay set at repling wal
}

type LeaderStatus struct {
	leader        uint64
	leaderChanged uint64
}

// newRaftServer initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func newRaftServer(hub *component.ComponentHub,
	cluster *Cluster,
	listenUrl string, join bool, snapdir string,
	certFile string, keyFile string,
	getSnapshot func() ([]byte, error),
	tickMS time.Duration,
	confChangeC <-chan raftpb.ConfChange,
	commitC chan *types.Block,
	delayPromote bool,
	chainWal consensus.ChainWAL) *raftServer {

	errorC := make(chan error)
	peers := cluster.BPUrls

	rs := &raftServer{
		ComponentHub: hub,
		cluster:      cluster,
		walDB:        NewWalDB(chainWal),
		confChangeC:  confChangeC,
		commitC:      commitC,
		errorC:       errorC,
		id:           cluster.ID,
		listenUrl:    listenUrl,
		peers:        peers,
		join:         join,
		snapdir:      snapdir,
		getSnapshot:  getSnapshot,
		snapCount:    defaultSnapCount,
		stopc:        make(chan struct{}),
		httpstopc:    make(chan struct{}),
		httpdonec:    make(chan struct{}),

		snapshotterReady: make(chan *snap.Snapshotter, 1),
		// rest of structure populated after WAL replay

		certFile: certFile,
		keyFile:  keyFile,

		lock:       sync.RWMutex{},
		promotable: true,
		tickMS:     tickMS,
	}

	if listenUrl == "" {
		rs.listenUrl = peers[rs.id-1]
	}

	if delayPromote {
		rs.SetPromotable(false)
	}

	return rs
}

func (rs *raftServer) SetPromotable(val bool) {
	defer rs.lock.Unlock()
	rs.lock.Lock()
	rs.promotable = val
}

func (rs *raftServer) GetPromotable() bool {
	defer rs.lock.RUnlock()

	rs.lock.RLock()
	val := rs.promotable

	return val
}

func (rs *raftServer) Start() {
	go rs.startRaft()
}

func (rs *raftServer) startRaft() {
	if !fileutil.Exist(rs.snapdir) {
		if err := os.MkdirAll(rs.snapdir, 0750); err != nil {
			logger.Error().Err(err).Msg("cannot create dir for snapshot")
		}
	}
	rs.snapshotter = snap.New(rs.snapdir)
	rs.snapshotterReady <- rs.snapshotter

	isNew := func() bool {
		return rs.walDB.IsNew()
	}

	if err := rs.replayWAL(); err != nil {
		logger.Fatal().Err(err).Msg("replay wal failed for raft")
	}

	rpeers := make([]raftlib.Peer, len(rs.peers))
	for i := range rpeers {
		rpeers[i] = raftlib.Peer{ID: uint64(i + 1)}
	}

	c := &raftlib.Config{
		ID:                        uint64(rs.id),
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   rs.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		Logger:                    raftLogger,
		CheckQuorum:               true,
		PreVote:                   true,
		DisableProposalForwarding: true,
	}

	var node raftlib.Node
	if !isNew() {
		node = raftlib.RestartNode(c)
	} else {
		startPeers := rpeers
		if rs.join {
			startPeers = nil
		}
		node = raftlib.StartNode(c, startPeers)
	}

	logger.Debug().Msg("raft core node is started")

	// need locking for sync with consensusAccessor
	rs.setNodeSync(node)

	rs.transport = &rafthttp.Transport{
		ID:          etcdtypes.ID(rs.id),
		ClusterID:   0x1000,
		Raft:        rs,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.FormatUint(rs.id, 10)),
		ErrorC:      make(chan error),
	}

	rs.transport.SetLogger(httpLogger)

	rs.transport.Start()
	for i := range rs.peers {
		if uint64(i+1) != rs.id {
			rs.transport.AddPeer(etcdtypes.ID(i+1), []string{rs.peers[i]})
		}
	}

	go rs.serveRaft()
	go rs.serveChannels()
}

func (rs *raftServer) setNodeSync(node raftlib.Node) {
	defer rs.lock.Unlock()

	rs.lock.Lock()
	rs.node = node
}

func (rs *raftServer) getNodeSync() raftlib.Node {
	defer rs.lock.RUnlock()

	var node raftlib.Node
	rs.lock.RLock()
	node = rs.node

	return node
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

// TODO timeout handling with context
func (rs *raftServer) Propose(block *types.Block) {
	if data, err := marshalEntryData(block); err == nil {
		// blocks until accepted by raft state machine
		rs.node.Propose(context.TODO(), data)
		logger.Debug().Int("len", len(data)).Msg("proposed data to raft node")
	} else {
		logger.Fatal().Err(err).Msg("poposed data is invalid")
	}
}

func (rs *raftServer) serveChannels() {
	snapshot, err := rs.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rs.confState = snapshot.Metadata.ConfState
	rs.snapshotIndex = snapshot.Metadata.Index
	rs.appliedIndex = snapshot.Metadata.Index

	ticker := time.NewTicker(rs.tickMS)
	defer ticker.Stop()

	// send proposals over raft
	go func() {
		var confChangeCount uint64 = 0

		for rs.confChangeC != nil {
			select {
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
			if rs.GetPromotable() {
				rs.node.Tick()
			}

			// store raft entries to walDB, then publish over commit channel
		case rd := <-rs.node.Ready():
			if len(rd.Entries) > 0 || len(rd.CommittedEntries) > 0 || !raftlib.IsEmptyHardState(rd.HardState) {
				logger.Debug().Int("entries", len(rd.Entries)).Int("commitentries", len(rd.CommittedEntries)).Str("term", rd.HardState.String()).Msg("ready to process")
			}

			rs.walDB.SaveEntry(rd.HardState, rd.Entries)
			if !raftlib.IsEmptySnap(rd.Snapshot) {
				if err := rs.saveSnap(&rd.Snapshot); err != nil {
					logger.Fatal().Err(err).Msg("failed to save snapshot")
					//TODO handle sync failure , need to retry(it can be occurred when leader stops while synchronizing)
				}

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

			// New block must be created after connecting all commited block
			if rd.SoftState != nil {
				rs.updateLeader(rd.SoftState)
			}

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
	urlstr := rs.listenUrl
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

func (rs *raftServer) loadSnapshot() (*raftpb.Snapshot, error) {
	snapshot, err := rs.walDB.GetSnapshot()
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading snapshot")
		return nil, err
	}

	if snapshot == nil {
		logger.Info().Msg("snapshot does not exist")
		return nil, nil
	}
	chainSnap, err := consensus.DecodeChainSnapshot(snapshot.Data)
	if err != nil {
		logger.Fatal().Err(err).Msg("error decoding snapshot")
		return nil, err
	}

	done, err := rs.walDB.GetSnapshotDone()
	if err != nil {
		logger.Fatal().Err(err).Msg("get snapshot status")
		return nil, err
	}

	logger.Info().Str("meta", snapshot.String()).Str("snap", chainSnap.ToString()).Msg("loaded snapshot meta")
	if done {
		return snapshot, nil
	}

	logger.Info().Str("meta", snapshot.String()).Str("snap", chainSnap.ToString()).Msg("need to sync with snapshot again")
	if err := rs.requestSync(chainSnap); err != nil {
		logger.Fatal().Err(err).Str("snap", chainSnap.ToString()).Msg("error sync snapshot")
		return nil, err
	}

	return snapshot, nil
}

/*
// openWAL returns a WAL ready for reading.
func (rs *raftServer) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rs.waldir) {
		if err := os.MkdirAll(rs.waldir, 0750); err != nil {
			logger.Fatal().Err(err).Msg("cannot create dir for walDB")
		}

		w, err := wal.Create(rs.waldir, nil)
		if err != nil {
			logger.Fatal().Err(err).Msg("create walDB error")
		}

		logger.Info().Str("dir", rs.waldir).Msg("create walDB directory")
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	logger.Info().Uint64("term", walsnap.Term).Uint64("index", walsnap.Index).Msg("loading WAL at term %d and index")
	w, err := wal.Open(rs.waldir, walsnap)
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading walDB")
	}

	logger.Info().Msg("openwal done")
	return w
}
*/

// replayWAL replays WAL entries into the raft instance.
func (rs *raftServer) replayWAL() error {
	logger.Info().Uint64("raftid", rs.id).Msg("replaying WAL of member %d")

	snapshot, err := rs.loadSnapshot()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
		return err
	}

	st, ents, err := rs.walDB.ReadAll(snapshot)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
		return err
	}

	rs.lastAppliedBlock, err = rs.walDB.GetBestBlock()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get last applied block")
		return err
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
		rs.startSync = true
	}

	logger.Info().Msg("replaying WAL done")

	return nil
}

func (rs *raftServer) createSnapshot(snapshotIndex uint64) ([]byte, error) {
	//get block of prev log
	if rs.lastAppliedBlock == nil {
		logger.Fatal().Msg("last applied block is nil")
	}

	snapBlockNo := rs.lastAppliedBlock.BlockNo() - 1
	snapBlock, err := rs.walDB.GetBlockByNo(snapBlockNo)
	if err != nil {
		logger.Fatal().Err(err).Uint64("no", snapBlockNo).Msg("failed to get prev of last applied block")
	}

	if !bytes.Equal(snapBlock.GetHash(), rs.lastAppliedBlock.GetHeader().GetPrevBlockHash()) {
		logger.Fatal().Err(err).Uint64("no", snapBlockNo).Str("snap", snapBlock.ID()).Str("prev of last", enc.ToString(rs.lastAppliedBlock.GetHeader().GetPrevBlockHash())).Msg("block in wal is not equal to raft")
	}

	logger.Info().Str("hash", snapBlock.ID()).Uint64("no", snapBlock.BlockNo()).Msg("create new snapshot")

	snap := consensus.NewChainSnapshot(snapBlock)
	if snap == nil {
		panic("new snap failed")
	}

	return snap.ToBytes()
}

// maybeTriggerSnapshot create snapshot and make compaction for raft log storage
// raft can not wait until last applied entry commits. so snapshot must create from rs.appliedIndex - 1
func (rs *raftServer) maybeTriggerSnapshot() {
	if rs.appliedIndex-rs.snapshotIndex <= rs.snapCount {
		return
	}

	newSnapshotIndex := rs.appliedIndex - 1

	logger.Info().Uint64("new snap index", newSnapshotIndex).Uint64("last snapshot index", rs.snapshotIndex).Msg("start snapshot")
	data, err := rs.createSnapshot(newSnapshotIndex)
	if err != nil {
		logger.Fatal().Err(err).Msg("raft getsnapshot failed")
	}
	_, err = rs.raftStorage.CreateSnapshot(newSnapshotIndex, &rs.confState, data)
	if err != nil {
		logger.Fatal().Err(err).Msg("raft create snapshot failed")
	}

	compactIndex := uint64(1)
	if newSnapshotIndex > snapshotCatchUpEntriesN {
		compactIndex = newSnapshotIndex - snapshotCatchUpEntriesN
	}
	if err := rs.raftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}

	logger.Info().Uint64("index", compactIndex).Msg("compacted raftLog.at index")
	rs.snapshotIndex = newSnapshotIndex
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

func (rs *raftServer) saveSnap(snap *raftpb.Snapshot) error {
	chainSnap, err := consensus.DecodeChainSnapshot(snap.Data)
	if err != nil {
		return err
	}

	// write snapshot log in WAL for crash recovery
	logger.Info().Str("snap", consensus.SnapToString(snap, chainSnap)).Msg("receive snapshot from remote")
	if err := rs.walDB.WriteSnapshot(snap); err != nil {
		return err
	}

	// TODO	request sync for chain with snapshot.data
	// wait to finish sync of chain
	if err := rs.requestSync(chainSnap); err != nil {
		logger.Fatal().Err(err).Msg("failed to sync. need to retry with other leader, try N times and shutdown")
		return err
	}

	return nil
}

// TODO handle error case that leader stops while synchronizing
func (rs *raftServer) requestSync(snap *consensus.ChainSnapshot) error {
	getSyncLeader := func() (peer.ID, error) {
		var peerID peer.ID
		var err error

		leader := rs.GetLeader()
		if leader == HasNoLeader {
			peerID, err = rs.cluster.getMemberPeerAddressToSync(snap.No)
			if err != nil {
				logger.Error().Err(err).Uint64("leader", leader).Msg("can't get peeraddress of leader")
				return "", err
			}
		} else {
			peerID, err = rs.cluster.getMemberPeerAddress(leader)
			if err != nil {
				logger.Error().Err(err).Uint64("leader", leader).Msg("can't get peeraddress of leader")
				return "", err
			}
		}

		logger.Debug().Str("peer", p2putil.ShortForm(peerID)).Msg("target peer to sync")

		return peerID, err
	}

	peerID, err := getSyncLeader()
	if err != nil {
		return err
	}

	if err := chain.SyncChain(rs.ComponentHub, snap.Hash, snap.No, peerID); err != nil {
		return err
	}

	if err := rs.walDB.WriteSnapshotDone(); err != nil {
		return err
	}

	return nil
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
			logger.Debug().Uint64("idx", ents[i].Index).Uint64("term", ents[i].Term).Int("datalen", len(ents[i].Data)).Msg("publish normal entry")

			var block *types.Block
			var err error
			if len(ents[i].Data) != 0 {
				if block, err = unmarshalEntryData(ents[i].Data); err != nil {
					logger.Fatal().Err(err).Uint64("idx", ents[i].Index).Uint64("term", ents[i].Term).Msg("commit entry is corrupted")
					continue
				}

				rs.lastAppliedBlock = block
			}
			select {
			case rs.commitC <- block:
			case <-rs.stopc:
				return false
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange

			cc.Unmarshal(ents[i].Data)
			rs.confState = *rs.node.ApplyConfChange(cc)

			logger.Debug().Int("idx", i).Str("type", raftpb.ConfChangeType_name[int32(cc.Type)]).Int("addrlen", len(cc.Context)).Msg("publish confchange entry")

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					rs.transport.AddPeer(etcdtypes.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rs.id) {
					logger.Info().Msg("I've been removed from the cluster! Shutting down.")
					return false
				}
				rs.transport.RemovePeer(etcdtypes.ID(cc.NodeID))
			}
		}

		// after commit, update appliedIndex
		rs.appliedIndex = ents[i].Index

		// special nil commit to signal replay has finished
		if ents[i].Index == rs.lastIndex {
			if !rs.startSync {
				logger.Debug().Uint64("idx", rs.lastIndex).Msg("published all entries of WAL")

				select {
				case rs.commitC <- nil:
					rs.startSync = true
				case <-rs.stopc:
					return false
				}
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

func (rs *raftServer) Status() raftlib.Status {
	node := rs.getNodeSync()
	if node == nil {
		return raftlib.Status{}
	}

	return node.Status()
}

func marshalEntryData(block *types.Block) ([]byte, error) {
	var data []byte
	var err error
	if data, err = proto.Marshal(block); err != nil {
		logger.Fatal().Err(err).Msg("poposed data is invalid")
	}

	return data, nil
}

var (
	ErrUnmarshal = errors.New("failed to unmarshalEntryData log entry")
)

func unmarshalEntryData(data []byte) (*types.Block, error) {
	block := &types.Block{}
	if err := proto.Unmarshal(data, block); err != nil {
		return block, ErrUnmarshal
	}

	return block, nil
}
