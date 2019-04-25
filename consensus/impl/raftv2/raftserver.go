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
	"encoding/binary"
	"errors"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/gogo/protobuf/proto"
	"io"
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
)

var (
	ErrNoSnapshot         = errors.New("no snapshot")
	ErrCCAlreadyApplied   = errors.New("conf change entry is already applied")
	ErrInvalidMember      = errors.New("member of conf change is invalid")
	ErrInvalidMemberID    = errors.New("member id of conf change doesn't match")
	ErrCCAlreadyAdded     = errors.New("member has already added")
	ErrCCNoMemberToRemove = errors.New("there is no member to remove")
)

const (
	HasNoLeader MemberID = 0
)

func init() {
	raftLogger = NewRaftLogger(logger)
}

// A key-value stream backed by raft
// A key-value stream backed by raft
type raftServer struct {
	*component.ComponentHub

	pa p2pcommon.PeerAccessor

	cluster *Cluster

	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan *types.Block        // entries committed to log (k,v)
	errorC      chan error               // errors from raft session

	id          MemberID // client ID for raft session
	listenUrl   string
	join        bool   // node is joining an existing cluster
	snapdir     string // path to snapshot directory
	getSnapshot func() ([]byte, error)
	lastIndex   uint64 // index of log at start

	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raftlib.Node
	raftStorage *raftlib.MemoryStorage
	//wal         *wal.WAL
	walDB *WalDB

	snapshotter      *ChainSnapshotter
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

	confState    raftpb.ConfState
	progress     BlockProgress
	prevProgress BlockProgress // prev state before appling last block
}

type BlockProgress struct {
	block     *types.Block //tracking last applied block. It's initillay set at repling wal
	index     uint64
	term      uint64
	confState raftpb.ConfState
}

func (prog *BlockProgress) isEmpty() bool {
	return prog.block == nil
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

	rs := &raftServer{
		ComponentHub: hub,
		cluster:      cluster,
		walDB:        NewWalDB(chainWal),
		confChangeC:  confChangeC,
		commitC:      commitC,
		errorC:       errorC,
		id:           cluster.NodeID,
		listenUrl:    listenUrl,
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

	if delayPromote {
		rs.SetPromotable(false)
	}

	return rs
}

func (rs *raftServer) SetPeerAccessor(pa p2pcommon.PeerAccessor) {
	rs.pa = pa
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

func (rs *raftServer) makeStartPeers() ([]raftlib.Peer, error) {
	rpeers := make([]raftlib.Peer, rs.cluster.Size)

	var i int
	for _, member := range rs.cluster.configMembers.MapByID {
		data, err := member.Marshal()
		if err != nil {
			return nil, err
		}
		rpeers[i] = raftlib.Peer{ID: uint64(member.ID), Context: data}
		i++
	}

	return rpeers, nil
}

func (rs *raftServer) startRaft() {
	if !fileutil.Exist(rs.snapdir) {
		if err := os.MkdirAll(rs.snapdir, 0750); err != nil {
			logger.Error().Err(err).Msg("cannot create dir for snapshot")
		}
	}
	rs.snapshotter = newChainSnapshotter(rs.pa, rs.ComponentHub, rs.cluster, rs.walDB, func() MemberID { return rs.GetLeader() })

	if err := rs.replayWAL(); err != nil {
		logger.Fatal().Err(err).Msg("replay wal failed for raft")
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
	last, err := rs.walDB.GetRaftEntryLastIdx()
	if err != nil {
		logger.Fatal().Msg("failed to get raft entry last index")
	}
	if last > 0 {
		logger.Info().Msg("raft restart")

		node = raftlib.RestartNode(c)
	} else {
		logger.Info().Msg("raft start at first time")

		rpeers, err := rs.makeStartPeers()
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to make raft peer list")
		}

		startPeers := rpeers
		if rs.join {
			startPeers = nil
		}
		node = raftlib.StartNode(c, startPeers)
	}

	logger.Debug().Msg("raft core node is started")

	// need locking for sync with consensusAccessor
	rs.setNodeSync(node)

	rs.startTransport()

	go rs.serveRaft()
	go rs.serveChannels()
}

func (rs *raftServer) startTransport() {
	rs.transport = &rafthttp.Transport{
		ID:          etcdtypes.ID(rs.id),
		ClusterID:   0x1000,
		Raft:        rs,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.FormatUint(uint64(rs.id), 10)),
		Snapshotter: rs.snapshotter,
		ErrorC:      make(chan error),
	}

	rs.transport.SetLogger(httpLogger)

	if err := rs.transport.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start raft http")
	}

	for _, member := range rs.cluster.getEffectiveMembers().MapByID {
		if rs.cluster.NodeID != member.ID {
			rs.transport.AddPeer(etcdtypes.ID(member.ID), []string{member.Url})
		}
	}
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
func (rs *raftServer) Propose(block *types.Block) error {
	if data, err := marshalEntryData(block); err == nil {
		// blocks until accepted by raft state machine
		if err := rs.node.Propose(context.TODO(), data); err != nil {
			return err
		}

		logger.Debug().Int("len", len(data)).Msg("proposed data to raft node")
	} else {
		logger.Fatal().Err(err).Msg("poposed data is invalid")
	}

	return nil
}

func (rs *raftServer) serveChannels() {
	snapshot, err := rs.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rs.setConfState(snapshot.Metadata.ConfState)
	rs.setSnapshotIndex(snapshot.Metadata.Index)
	rs.setAppliedIndex(snapshot.Metadata.Index)

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
					if err := rs.node.ProposeConfChange(context.TODO(), cc); err != nil {
						logger.Fatal().Err(err).Msg("failed to propose configure change")
					}
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
				logger.Debug().Int("entries", len(rd.Entries)).Int("commitentries", len(rd.CommittedEntries)).Str("hardstate", rd.HardState.String()).Msg("ready to process")
			}

			if rs.IsLeader() {
				if err := rs.processMessages(rd.Messages); err != nil {
					logger.Fatal().Err(err).Msg("leader process message error")
				}
			}

			if err := rs.walDB.SaveEntry(rd.HardState, rd.Entries); err != nil {
				logger.Fatal().Err(err).Msg("failed to save entry to wal")
			}

			if !raftlib.IsEmptySnap(rd.Snapshot) {
				if err := rs.raftStorage.ApplySnapshot(rd.Snapshot); err != nil {
					logger.Fatal().Err(err).Msg("failed to apply snapshot")
				}

				rs.publishSnapshot(rd.Snapshot)
			}
			if err := rs.raftStorage.Append(rd.Entries); err != nil {
				logger.Fatal().Err(err).Msg("failed to append new entries to raft log")
			}

			if !rs.IsLeader() {
				if err := rs.processMessages(rd.Messages); err != nil {
					logger.Fatal().Err(err).Msg("process message error")
				}
			}
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

func (rs *raftServer) processMessages(msgs []raftpb.Message) error {
	var err error
	var tmpSnapMsg *snap.Message

	snapMsgs := make([]*snap.Message, 0)

	// reset MsgSnap to send snap.Message
	for i, msg := range msgs {
		if msg.Type == raftpb.MsgSnap {
			tmpSnapMsg, err = rs.makeSnapMessage(&msg)
			if err != nil {
				return err
			}
			snapMsgs = append(snapMsgs, tmpSnapMsg)

			msgs[i].To = 0
		}
	}

	rs.transport.Send(msgs)

	for _, tmpSnapMsg := range snapMsgs {
		rs.transport.SendSnapshot(*tmpSnapMsg)
	}

	return nil
}

func (rs *raftServer) makeSnapMessage(msg *raftpb.Message) (*snap.Message, error) {
	if msg.Type != raftpb.MsgSnap {
		return nil, ErrNotMsgSnap
	}

	/*
		// make snapshot with last progress of raftserver
		snapshot, err := rs.snapshotter.createSnapshot(rs.prevProgress, rs.confState)
		if err != nil {
			return nil, err
		}

		msg.Snapshot = *snapshot
	*/
	// TODO add cluster info to snapshot.data

	logger.Debug().Uint64("term", msg.Term).Uint64("index", msg.Index).Msg("send merged snapshot message")

	// not using pipe to send snapshot
	pr, pw := io.Pipe()

	go func() {
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, int32(1))
		if err != nil {
			logger.Fatal().Err(err).Msg("raft pipe binary write err")
		}

		n, err := pw.Write(buf.Bytes())
		if err == nil {
			logger.Debug().Msgf("wrote database snapshot out [total bytes: %d]", n)
		} else {
			logger.Error().Msgf("failed to write database snapshot out [written bytes: %d]: %v", n, err)
		}
		if err := pw.CloseWithError(err); err != nil {
			logger.Fatal().Err(err).Msg("raft pipe close error")
		}
	}()

	return snap.NewMessage(*msg, pr, 4), nil
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

	logger.Info().Str("meta", consensus.SnapToString(snapshot, chainSnap)).Msg("loaded snapshot meta")

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
func (rs *raftServer) updateBlockProgress(term uint64, index uint64, block *types.Block) {
	if block == nil {
		return
	}

	logger.Debug().Uint64("term", term).Uint64("index", index).Uint64("no", block.BlockNo()).Str("hash", block.ID()).Msg("set progress of last block")

	rs.prevProgress = rs.progress

	rs.progress.term = term
	rs.progress.index = index
	rs.progress.block = block
	rs.progress.confState = rs.confState
}

// replayWAL replays WAL entries into the raft instance.
func (rs *raftServer) replayWAL() error {
	logger.Info().Uint64("raftid", uint64(rs.id)).Msg("replaying WAL")

	snapshot, err := rs.loadSnapshot()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
		return err
	}

	// TODO recover cluster from snapshot

	st, ents, err := rs.walDB.ReadAll(snapshot)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
		return err
	}

	rs.raftStorage = raftlib.NewMemoryStorage()
	if snapshot != nil {
		if err := rs.raftStorage.ApplySnapshot(*snapshot); err != nil {
			logger.Fatal().Err(err).Msg("failed to apply snapshot to reaply wal")
		}
	}
	if err := rs.raftStorage.SetHardState(*st); err != nil {
		logger.Fatal().Err(err).Msg("failed to set hard state to reaply wal")
	}

	// append to storage so raft starts at the right place in log
	if err := rs.raftStorage.Append(ents); err != nil {
		logger.Fatal().Err(err).Msg("failed to append entries to reaply wal")
	}
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

/*
// createSnapshot make marshalled data of chain & cluster info
func (rs *raftServer) createSnapshot() ([]byte, error) {
	// this snapshot is used when reboot and initialize raft log
	if rs.prevProgress.isEmpty() {
		logger.Fatal().Msg("last applied block is nil")
	}

	snapBlock := rs.prevProgress.block

	logger.Info().Str("hash", snapBlock.ID()).Uint64("no", snapBlock.BlockNo()).Msg("create new snapshot of chain")

	snap := consensus.NewChainSnapshot(snapBlock)
	if snap == nil {
		panic("new snap failed")
	}

	return snap.ToBytes()
}*/

// maybeTriggerSnapshot create snapshot and make compaction for raft log storage
// raft can not wait until last applied entry commits. so snapshot must create from rs.prevProgress.index
func (rs *raftServer) maybeTriggerSnapshot() {
	if rs.prevProgress.index == 0 || rs.prevProgress.block == nil {
		logger.Debug().Msg("raft need to make snapshot, but does not exist target progress")
		return
	}

	newSnapshotIndex := rs.prevProgress.index

	if newSnapshotIndex-rs.snapshotIndex <= rs.snapCount {
		return
	}

	logger.Info().Uint64("applied", rs.appliedIndex).Uint64("new snap index", newSnapshotIndex).Uint64("last snapshot index", rs.snapshotIndex).Msg("start snapshot")

	// make snapshot data of previous connected block
	chainsnap, err := rs.snapshotter.createSnapshotData(rs.prevProgress.block)
	if err != nil {
		logger.Fatal().Msg("failed to create snapshot data from prev block")
	}

	data, err := chainsnap.ToBytes()

	// snapshot.data is not used for snapshot transfer. At the time of transmission, a message is generated again with information at that time and sent.
	snapshot, err := rs.raftStorage.CreateSnapshot(newSnapshotIndex, &rs.prevProgress.confState, data)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create snapshot")
	}

	// save snapshot to wal
	if err := rs.walDB.WriteSnapshot(&snapshot, true); err != nil {
		logger.Fatal().Err(err).Msg("failed to write snapshot")
	}

	compactIndex := uint64(1)
	if newSnapshotIndex > snapshotCatchUpEntriesN {
		compactIndex = newSnapshotIndex - snapshotCatchUpEntriesN
	}
	if err := rs.raftStorage.Compact(compactIndex); err != nil {
		if err == raftlib.ErrCompacted {
			return
		}
		panic(err)
	}

	logger.Info().Uint64("index", compactIndex).Msg("compacted raftLog.at index")
	rs.setSnapshotIndex(newSnapshotIndex)
}

func (rs *raftServer) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raftlib.IsEmptySnap(snapshotToSave) {
		return
	}

	logger.Info().Uint64("index", rs.snapshotIndex).Str("snap", consensus.SnapToString(&snapshotToSave, nil)).Msg("publishing snapshot at index")
	defer logger.Info().Uint64("index", rs.snapshotIndex).Msg("finished publishing snapshot at index")

	if snapshotToSave.Metadata.Index <= rs.appliedIndex {
		logger.Fatal().Msgf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index, rs.appliedIndex)
	}
	//rs.commitC <- nil // trigger kvstore to load snapshot

	rs.setConfState(snapshotToSave.Metadata.ConfState)
	rs.setSnapshotIndex(snapshotToSave.Metadata.Index)
	rs.setAppliedIndex(snapshotToSave.Metadata.Index)

	rs.prevProgress.index = 0
	rs.progress.index = 0

	// TODO reconnect peers
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

var (
	ErrInvCCType = errors.New("change type of ")
)

func (rs *raftServer) validateConfChange(cc *raftpb.ConfChange, member *Member) error {
	if !member.isValid() {
		return ErrInvalidMember
	}

	if cc.NodeID != uint64(member.ID) {
		return ErrInvalidMemberID
	}

	cluster := rs.cluster
	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		if m := cluster.members.getMember(member.ID); m != nil {
			return ErrCCAlreadyAdded
		}

		if err := cluster.members.hasDuplicatedMember(member); err != nil {
			return err
		}

	case raftpb.ConfChangeRemoveNode:
		if m := cluster.members.getMember(member.ID); m == nil {
			return ErrCCNoMemberToRemove
		}
	default:
		return ErrInvCCType
	}

	// - TODO UPDATE
	return nil
}

func (rs *raftServer) ValidateConfChangeEntry(entry *raftpb.Entry) (*raftpb.ConfChange, *Member, error) {
	// TODO XXX validate from current cluster configure
	var cc *raftpb.ConfChange
	var member *Member
	var err error

	alreadyApplied := func(entry *raftpb.Entry) bool {
		return rs.cluster.appliedTerm >= entry.Term || rs.cluster.appliedIndex >= entry.Index
	}

	if alreadyApplied(entry) {
		return nil, nil, ErrCCAlreadyApplied
	}

	unmarshalConfChangeEntry := func() (*raftpb.ConfChange, *Member, error) {
		var cc raftpb.ConfChange

		if err := cc.Unmarshal(entry.Data); err != nil {
			logger.Fatal().Err(err).Uint64("idx", entry.Index).Uint64("term", entry.Term).Msg("failed to unmarshal of conf change entry")
			return nil, nil, err
		}

		// skip confchange of empty context
		if len(cc.Context) == 0 {
			return nil, nil, nil
		}

		var member = &Member{}
		if err := member.Unmarshal(cc.Context); err != nil {
			logger.Fatal().Err(err).Uint64("idx", entry.Index).Uint64("term", entry.Term).Msg("failed to unmarshal of context of cc entry")
			return nil, nil, err
		}

		return &cc, member, nil
	}

	cc, member, err = unmarshalConfChangeEntry()

	if err = rs.validateConfChange(cc, member); err != nil {
		return cc, member, err
	}

	return cc, member, nil
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rs *raftServer) publishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		logger.Info().Uint64("idx", ents[i].Index).Uint64("term", ents[i].Term).Str("type", ents[i].Type.String()).Int("datalen", len(ents[i].Data)).Msg("publish entry")

		switch ents[i].Type {
		case raftpb.EntryNormal:
			var block *types.Block
			var err error
			if len(ents[i].Data) != 0 {
				if block, err = unmarshalEntryData(ents[i].Data); err != nil {
					logger.Fatal().Err(err).Uint64("idx", ents[i].Index).Uint64("term", ents[i].Term).Msg("commit entry is corrupted")
					continue
				}

			}

			if block != nil {
				logger.Info().Str("hash", block.ID()).Uint64("no", block.BlockNo()).Msg("commit normal block entry")
			}

			select {
			case rs.commitC <- block:
			case <-rs.stopc:
				return false
			}
			rs.updateBlockProgress(ents[i].Term, ents[i].Index, block)

		case raftpb.EntryConfChange:
			var cc *raftpb.ConfChange
			var member *Member
			var err error

			if cc, member, err = rs.ValidateConfChangeEntry(&ents[i]); err != nil {
				logger.Warn().Err(err).Msg("failed to validate conf change")
				// reset pending conf change
				cc.NodeID = raftlib.None
				rs.node.ApplyConfChange(*cc)
			} else {
				rs.confState = *rs.node.ApplyConfChange(*cc)

				logger.Info().Str("type", cc.Type.String()).Str("member", member.ToString()).Msg("publish confchange entry")

				switch cc.Type {
				case raftpb.ConfChangeAddNode:
					if len(cc.Context) > 0 {
						rs.transport.AddPeer(etcdtypes.ID(cc.NodeID), []string{member.Url})
					}
				case raftpb.ConfChangeRemoveNode:
					if cc.NodeID == uint64(rs.id) {
						logger.Info().Msg("I've been removed from the cluster! Shutting down.")
						return false
					}
					rs.transport.RemovePeer(etcdtypes.ID(cc.NodeID))
				}
			}
		}

		// after commit, update appliedIndex
		rs.setAppliedIndex(ents[i].Index)

		/* XXX no need commitC <- nil
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
		}*/
	}
	return true
}

func (rs *raftServer) setSnapshotIndex(idx uint64) {
	logger.Debug().Uint64("index", idx).Msg("raft server set snapshotIndex")

	rs.snapshotIndex = idx
}

func (rs *raftServer) setAppliedIndex(idx uint64) {
	logger.Debug().Uint64("index", idx).Msg("raft server set appliedIndex")

	rs.appliedIndex = idx
}

func (rs *raftServer) setConfState(state raftpb.ConfState) {
	logger.Debug().Str("state", consensus.ConfStateToString(&state)).Msg("raft server set confstate")

	rs.confState = state
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
	if MemberID(softState.Lead) != rs.GetLeader() {
		atomic.StoreUint64(&rs.leaderStatus.leader, softState.Lead)

		rs.leaderStatus.leaderChanged++

		logger.Info().Str("ID", MemberIDToString(rs.id)).Uint64("leader", softState.Lead).Msg("leader changed")
	}
}

func (rs *raftServer) GetLeader() MemberID {
	return MemberID(atomic.LoadUint64(&rs.leaderStatus.leader))
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
