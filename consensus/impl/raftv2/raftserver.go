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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/gogo/protobuf/proto"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"

	"github.com/aergoio/etcd/etcdserver/stats"
	etcdtypes "github.com/aergoio/etcd/pkg/types"
	raftlib "github.com/aergoio/etcd/raft"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/aergoio/etcd/rafthttp"
	"github.com/aergoio/etcd/snap"
)

const (
	HasNoLeader       uint64 = 0
	DfltSnapFrequency        = 30
)

//noinspection ALL
var (
	raftLogger                  raftlib.Logger
	ConfSnapFrequency           uint64 = DfltSnapFrequency
	ConfSnapshotCatchUpEntriesN uint64 = ConfSnapFrequency

	MaxProgressDiff uint64 = 100
)

var (
	ErrCCAlreadyApplied    = errors.New("conf change entry is already applied")
	ErrInvalidMember       = errors.New("member of conf change is invalid")
	ErrCCAlreadyAdded      = errors.New("member has already added")
	ErrCCAlreadyRemoved    = errors.New("member has already removed")
	ErrCCNoMemberToRemove  = errors.New("there is no member to remove")
	ErrEmptySnapshot       = errors.New("received empty snapshot")
	ErrInvalidRaftIdentity = errors.New("raft identity is not set")
	ErrProposeNilBlock     = errors.New("proposed block is nil")
)

func init() {
	raftLogger = NewRaftLogger(logger)
}

// A key-value stream backed by raft
// A key-value stream backed by raft
type raftServer struct {
	*component.ComponentHub
	sync.RWMutex

	pa p2pcommon.PeerAccessor

	cluster *Cluster

	confChangeC <-chan *consensus.ConfChangePropose // proposed cluster config changes
	commitC     chan *commitEntry                   // entries committed to log (k,v)
	errorC      chan error                          // errors from raft session

	id              uint64 // client ID for raft session
	listenUrl       string
	join            bool // node is joining an existing cluster
	joinUsingBackup bool // use backup chaindb datafiles to join a existing cluster
	getSnapshot     func() ([]byte, error)
	lastIndex       uint64 // index of log at start

	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raftlib.Node
	raftStorage *raftlib.MemoryStorage
	//wal         *wal.WAL
	walDB *WalDB

	snapshotter *ChainSnapshotter

	snapFrequency uint64
	transport     *rafthttp.Transport
	stopc         chan struct{} // signals proposal channel closed
	httpstopc     chan struct{} // signals http server to shutdown
	httpdonec     chan struct{} // signals http server shutdown complete

	leaderStatus LeaderStatus

	certFile string
	keyFile  string

	promotable bool

	tickMS time.Duration

	confState *raftpb.ConfState

	commitProgress CommitProgress
}

type LeaderStatus struct {
	leader        uint64
	leaderChanged uint64
}

type commitEntry struct {
	block *types.Block
	index uint64
	term  uint64
}

type CommitProgress struct {
	sync.Mutex

	connect commitEntry // last connected entry to chain
	request commitEntry // last requested entry to commitC
}

func (cp *CommitProgress) UpdateConnect(ce *commitEntry) {
	logger.Debug().Uint64("term", ce.term).Uint64("index", ce.index).Uint64("no", ce.block.BlockNo()).Str("hash", ce.block.ID()).Msg("set progress of last connected block")

	cp.Lock()
	defer cp.Unlock()

	cp.connect = *ce
}

func (cp *CommitProgress) UpdateRequest(ce *commitEntry) {
	logger.Debug().Uint64("term", ce.term).Uint64("index", ce.index).Uint64("no", ce.block.BlockNo()).Str("hash", ce.block.ID()).Msg("set progress of last request block")

	cp.Lock()
	defer cp.Unlock()

	cp.request = *ce
}

func (cp *CommitProgress) GetConnect() *commitEntry {
	cp.Lock()
	defer cp.Unlock()

	return &cp.connect
}

func (cp *CommitProgress) GetRequest() *commitEntry {
	cp.Lock()
	defer cp.Unlock()

	return &cp.request
}

func (cp *CommitProgress) IsReadyToPropose() bool {
	cp.Lock()
	defer cp.Unlock()

	if cp.request.block == nil {
		return true
	}

	var connNo, reqNo uint64
	reqNo = cp.request.block.BlockNo()
	if cp.connect.block != nil {
		connNo = cp.connect.block.BlockNo()
	}

	if reqNo <= connNo {
		return true
	}

	logger.Debug().Uint64("requested", reqNo).Uint64("connected", connNo).Msg("remain pending request to connect")

	return false
}

func RecoverExit() {
	if r := recover(); r != nil {
		logger.Error().Str("callstack", string(debug.Stack())).Msg("panic occurred in raft server")
		os.Exit(10)
	}
}

func makeConfig(nodeID uint64, storage *raftlib.MemoryStorage) *raftlib.Config {
	c := &raftlib.Config{
		ID:                        nodeID,
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   storage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		Logger:                    raftLogger,
		CheckQuorum:               true,
		DisableProposalForwarding: true,
	}

	return c
}

// newRaftServer initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func newRaftServer(hub *component.ComponentHub,
	cluster *Cluster,
	listenUrl string, join bool, useBackup bool,
	certFile string, keyFile string,
	getSnapshot func() ([]byte, error),
	tickMS time.Duration,
	confChangeC chan *consensus.ConfChangePropose,
	commitC chan *commitEntry,
	delayPromote bool,
	chainWal consensus.ChainWAL) *raftServer {

	errorC := make(chan error, 1)

	rs := &raftServer{
		ComponentHub:    hub,
		RWMutex:         sync.RWMutex{},
		cluster:         cluster,
		walDB:           NewWalDB(chainWal),
		confChangeC:     confChangeC,
		commitC:         commitC,
		errorC:          errorC,
		listenUrl:       listenUrl,
		join:            join,
		joinUsingBackup: useBackup,
		getSnapshot:     getSnapshot,
		snapFrequency:   ConfSnapFrequency,
		stopc:           make(chan struct{}),
		httpstopc:       make(chan struct{}),
		httpdonec:       make(chan struct{}),

		// rest of structure populated after WAL replay

		certFile: certFile,
		keyFile:  keyFile,

		promotable:     true,
		tickMS:         tickMS,
		commitProgress: CommitProgress{},
	}

	if delayPromote {
		rs.SetPromotable(false)
	}

	rs.snapshotter = newChainSnapshotter(nil, rs.ComponentHub, rs.cluster, rs.walDB, func() uint64 { return rs.GetLeader() })

	return rs
}

func (rs *raftServer) SetPeerAccessor(pa p2pcommon.PeerAccessor) {
	rs.pa = pa
	rs.snapshotter.setPeerAccessor(pa)
}

func (rs *raftServer) SetPromotable(val bool) {
	rs.Lock()
	defer rs.Unlock()

	rs.promotable = val
}

func (rs *raftServer) GetPromotable() bool {
	rs.RLock()
	defer rs.RUnlock()

	val := rs.promotable

	return val
}

func (rs *raftServer) Start() {
	go rs.startRaft()
}

func (rs *raftServer) makeStartPeers() ([]raftlib.Peer, error) {
	return rs.cluster.getStartPeers()
}

type RaftServerState int

const (
	RaftServerStateRestart = iota
	RaftServerStateNewCluster
	RaftServerStateJoinCluster
)

func (rs *raftServer) startRaft() {
	var node raftlib.Node

	getState := func() RaftServerState {
		hasWal, err := rs.walDB.HasWal(rs.cluster.identity)
		if err != nil {
			logger.Info().Msg("wal database of raft is missing or not valid")
		}

		switch {
		case hasWal:
			return RaftServerStateRestart
		case rs.join:
			return RaftServerStateJoinCluster
		default:
			return RaftServerStateNewCluster
		}

	}

	switch getState() {
	case RaftServerStateRestart:
		logger.Info().Msg("raft restart from wal")

		rs.cluster.ResetMembers()

		node = rs.restartNode()

	case RaftServerStateJoinCluster:
		logger.Info().Msg("raft start and join existing cluster")

		// get cluster info from existing cluster member and hardstate of bestblock
		existCluster, hardstateinfo, err := rs.GetExistingCluster()
		if err != nil {
			logger.Fatal().Err(err).Str("mine", rs.cluster.toString()).Msg("failed to get existing cluster info")
		}

		// config validate
		if !rs.cluster.ValidateAndMergeExistingCluster(existCluster) {
			logger.Fatal().Str("existcluster", existCluster.toString()).Str("mycluster", rs.cluster.toString()).Msg("this cluster configuration is not compatible with existing cluster")
		}

		if rs.joinUsingBackup {
			logger.Info().Msg("raft use given backup as wal")

			if err := rs.walDB.ResetWAL(hardstateinfo); err != nil {
				logger.Fatal().Err(err).Msg("reset wal failed for raft")
			}

			if err := rs.SaveIdentity(); err != nil {
				logger.Fatal().Err(err).Msg("fafiled to save identity")
			}

			node = rs.restartNode()

			logger.Info().Msg("raft restarted from backup")
		} else {
			node = rs.startNode(nil)
		}
	case RaftServerStateNewCluster:
		logger.Info().Msg("raft start and make new cluster")

		var startPeers []raftlib.Peer

		startPeers, err := rs.makeStartPeers()
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to make raft peer list")
		}

		node = rs.startNode(startPeers)
	}

	// need locking for sync with consensusAccessor
	rs.setNodeSync(node)

	rs.startTransport()

	go rs.serveRaft()
	go rs.serveChannels()
}

func (rs *raftServer) ID() uint64 {
	return rs.cluster.NodeID()
}

func (rs *raftServer) startNode(startPeers []raftlib.Peer) raftlib.Node {
	var (
		blk *types.Block
		err error
	)
	validateEmpty := func() {
		if blk, err = rs.walDB.GetBestBlock(); err != nil {
			logger.Fatal().Err(err).Msg("failed to get best block, so failed to start raft server")
		}
		if blk.BlockNo() > 0 {
			logger.Fatal().Err(err).Msg("blockchain data is not empty, so failed to start raft server")
		}
	}

	validateEmpty()

	if err := rs.cluster.SetThisNodeID(); err != nil {
		logger.Fatal().Err(err).Msg("failed to set id of this node")
	}

	// when join to exising cluster, cluster ID is already set
	if rs.cluster.ClusterID() == InvalidClusterID {
		rs.cluster.GenerateID()
	}

	if err := rs.SaveIdentity(); err != nil {
		logger.Fatal().Err(err).Str("identity", rs.cluster.identity.ToString()).Msg("failed to save identity")
	}

	rs.raftStorage = raftlib.NewMemoryStorage()

	c := makeConfig(rs.ID(), rs.raftStorage)

	logger.Info().Msg("raft node start")

	return raftlib.StartNode(c, startPeers)
}

func (rs *raftServer) restartNode() raftlib.Node {
	snapshot, err := rs.loadSnapshot()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read snapshot")
	}

	if err := rs.replayWAL(snapshot); err != nil {
		logger.Fatal().Err(err).Msg("replay wal failed for raft")
	}

	// members of cluster will be loaded from snapshot or wal
	if snapshot != nil {
		if err := rs.cluster.Recover(snapshot); err != nil {
			logger.Fatal().Err(err).Msg("failt to recover cluster from snapshot")
		}
	}

	c := makeConfig(rs.ID(), rs.raftStorage)

	logger.Info().Msg("raft node restart")

	return raftlib.RestartNode(c)
}

func (rs *raftServer) startTransport() {
	rs.transport = &rafthttp.Transport{
		ID:          etcdtypes.ID(rs.ID()),
		ClusterID:   etcdtypes.ID(rs.cluster.ClusterID()),
		Raft:        rs,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.FormatUint(rs.ID(), 10)),
		Snapshotter: rs.snapshotter,
		ErrorC:      rs.errorC,
	}

	rs.transport.SetLogger(httpLogger)

	if err := rs.transport.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start raft http")
	}

	for _, member := range rs.cluster.Members().MapByID {
		if rs.cluster.NodeID() != member.ID {
			rs.transport.AddPeer(etcdtypes.ID(member.ID), []string{member.Url})
		}
	}
}

func (rs *raftServer) SaveIdentity() error {
	if rs.cluster.ClusterID() == 0 || rs.cluster.NodeID() == consensus.InvalidMemberID || len(rs.cluster.NodeName()) == 0 {
		return ErrInvalidRaftIdentity
	}

	if err := rs.walDB.WriteIdentity(&rs.cluster.identity); err != nil {
		logger.Fatal().Err(err).Msg("failed to write raft identity to wal")
		return err
	}

	return nil
}

func (rs *raftServer) setNodeSync(node raftlib.Node) {
	rs.Lock()
	defer rs.Unlock()

	rs.node = node
}

func (rs *raftServer) getNodeSync() raftlib.Node {
	var node raftlib.Node

	rs.RLock()
	defer rs.RUnlock()

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
	logger.Error().Err(err).Msg("write err has occurend raft server. ")
}

// TODO timeout handling with context
func (rs *raftServer) Propose(block *types.Block) error {
	if block == nil {
		return ErrProposeNilBlock
	}
	logger.Debug().Msg("propose block")

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

func (rs *raftServer) saveConfChangeState(id uint64, state types.ConfChangeState, errCC error) error {
	var errStr string

	if errCC != nil {
		errStr = errCC.Error()
	}

	pr := types.ConfChangeProgress{State: state, Err: errStr}

	return rs.walDB.WriteConfChangeProgress(id, &pr)
}

func (rs *raftServer) serveConfChange() {
	handleConfChange := func(propose *consensus.ConfChangePropose) {
		var err error

		err = rs.node.ProposeConfChange(context.TODO(), *propose.Cc)
		if err != nil {
			logger.Error().Err(err).Msg("failed to propose configure change")
			rs.cluster.AfterConfChange(propose.Cc, nil, err)
		}

		err = rs.saveConfChangeState(propose.Cc.ID, types.ConfChangeState_CONF_CHANGE_STATE_PROPOSED, err)
		if err != nil {
			logger.Error().Err(err).Msg("failed to save progress of configure change")
		}
	}

	// send proposals over raft
	for rs.confChangeC != nil {
		select {
		case confChangePropose, ok := <-rs.confChangeC:
			if !ok {
				rs.confChangeC = nil
			} else {
				handleConfChange(confChangePropose)
			}
		}
	}
	// client closed channel; shutdown raft if not already
	close(rs.stopc)
}

func (rs *raftServer) serveChannels() {
	defer RecoverExit()

	snapshot, err := rs.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rs.setConfState(&snapshot.Metadata.ConfState)
	rs.setSnapshotIndex(snapshot.Metadata.Index)
	rs.setAppliedIndex(snapshot.Metadata.Index)

	ticker := time.NewTicker(rs.tickMS)
	defer ticker.Stop()

	go rs.serveConfChange()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			if rs.GetPromotable() {
				rs.node.Tick()
			}

			// store raft entries to walDB, then publish over commit channel
		case rd := <-rs.node.Ready():
			if len(rd.Entries) > 0 || len(rd.CommittedEntries) > 0 || !raftlib.IsEmptyHardState(rd.HardState) || rd.SoftState != nil {
				logger.Debug().Int("entries", len(rd.Entries)).Int("commitentries", len(rd.CommittedEntries)).Str("hardstate", types.RaftHardStateToString(rd.HardState)).Msg("ready to process")
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
				if err := rs.walDB.WriteSnapshot(&rd.Snapshot); err != nil {
					logger.Fatal().Err(err).Msg("failed to save snapshot to wal")
				}

				if err := rs.raftStorage.ApplySnapshot(rd.Snapshot); err != nil {
					logger.Fatal().Err(err).Msg("failed to apply snapshot")
				}

				if err := rs.publishSnapshot(rd.Snapshot); err != nil {
					logger.Fatal().Err(err).Msg("failed to publish snapshot")
				}
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
			rs.triggerSnapshot()

			// New block must be created after connecting all commited block
			if rd.SoftState != nil {
				rs.updateLeader(rd.SoftState)
			}

			rs.node.Advance()
		case err := <-rs.errorC:
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
	defer RecoverExit()

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

	snapdata := &consensus.SnapshotData{}

	err = snapdata.Decode(snapshot.Data)
	if err != nil {
		logger.Fatal().Err(err).Msg("error decoding snapshot")
		return nil, err
	}

	logger.Info().Str("meta", consensus.SnapToString(snapshot, snapdata)).Msg("loaded snapshot meta")

	return snapshot, nil
}

// replayWAL replays WAL entries into the raft instance.
func (rs *raftServer) replayWAL(snapshot *raftpb.Snapshot) error {
	logger.Info().Msg("replaying WAL")

	identity, st, ents, err := rs.walDB.ReadAll(snapshot)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read WAL")
		return err
	}

	if err := rs.cluster.RecoverIdentity(identity); err != nil {
		logger.Fatal().Err(err).Msg("failed to recover raft identity from wal")
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
	}

	logger.Info().Uint64("lastindex", rs.lastIndex).Msg("replaying WAL done")

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

	return snap.Encode()
}*/

// triggerSnapshot create snapshot and make compaction for raft log storage
// raft can not wait until last applied entry commits. so snapshot must create from current best block.
//
// @ MatchBlockAndCluster
// 	snapshot use current state of cluster and confstate. but last applied block may not be commited yet.
// 	so raft use last commited block. because of this, some conf change log can cause error on node that received snapshot
func (rs *raftServer) triggerSnapshot() {
	ce := rs.commitProgress.GetConnect()
	newSnapshotIndex, snapBlock := ce.index, ce.block

	if newSnapshotIndex == 0 || rs.confState == nil {
		return
	}

	if len(rs.confState.Nodes) == 0 {
		// TODO Fatal -> Error after test
		logger.Fatal().Msg("confstate node is empty for snapshot")
		return
	}

	if newSnapshotIndex-rs.snapshotIndex <= rs.snapFrequency {
		return
	}

	logger.Info().Uint64("applied", rs.appliedIndex).Uint64("new snap index", newSnapshotIndex).Uint64("last snapshot index", rs.snapshotIndex).Msg("start snapshot")

	// make snapshot data of best block
	snapdata, err := rs.snapshotter.createSnapshotData(rs.cluster, snapBlock, rs.confState)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create snapshot data from prev block")
	}

	data, err := snapdata.Encode()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to marshal snapshot data")
	}

	// snapshot.data is not used for snapshot transfer. At the time of transmission, a message is generated again with information at that time and sent.
	snapshot, err := rs.raftStorage.CreateSnapshot(newSnapshotIndex, rs.confState, data)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create snapshot")
	}

	// save snapshot to wal
	if err := rs.walDB.WriteSnapshot(&snapshot); err != nil {
		logger.Fatal().Err(err).Msg("failed to write snapshot")
	}

	compactIndex := uint64(1)
	if newSnapshotIndex > ConfSnapshotCatchUpEntriesN {
		compactIndex = newSnapshotIndex - ConfSnapshotCatchUpEntriesN
	}
	if err := rs.raftStorage.Compact(compactIndex); err != nil {
		if err == raftlib.ErrCompacted {
			return
		}
		panic(err)
	}

	logger.Info().Uint64("index", compactIndex).Msg("compacted raftLog.at index")
	rs.setSnapshotIndex(newSnapshotIndex)

	_ = chain.TestDebugger.Check(chain.DEBUG_RAFT_SNAP_FREQ, 0,
		func(freq int) error {
			rs.snapFrequency = uint64(freq)
			return nil
		})
}

func (rs *raftServer) publishSnapshot(snapshotToSave raftpb.Snapshot) error {
	updateProgress := func() error {
		var snapdata = &consensus.SnapshotData{}

		err := snapdata.Decode(snapshotToSave.Data)
		if err != nil {
			logger.Error().Msg("failed to unmarshal snapshot data to progress")
			return err
		}

		block, err := rs.walDB.GetBlockByNo(snapdata.Chain.No)
		if err != nil {
			logger.Fatal().Msg("failed to get synchronized block")
			return err
		}

		rs.commitProgress.UpdateConnect(&commitEntry{block: block, index: snapshotToSave.Metadata.Index, term: snapshotToSave.Metadata.Term})

		return nil
	}

	if raftlib.IsEmptySnap(snapshotToSave) {
		return ErrEmptySnapshot
	}

	logger.Info().Uint64("index", rs.snapshotIndex).Str("snap", consensus.SnapToString(&snapshotToSave, nil)).Msg("publishing snapshot at index")
	defer logger.Info().Uint64("index", rs.snapshotIndex).Msg("finished publishing snapshot at index")

	if snapshotToSave.Metadata.Index <= rs.appliedIndex {
		logger.Fatal().Msgf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index, rs.appliedIndex)
	}
	//rs.commitC <- nil // trigger kvstore to load snapshot

	rs.setConfState(&snapshotToSave.Metadata.ConfState)
	rs.setSnapshotIndex(snapshotToSave.Metadata.Index)
	rs.setAppliedIndex(snapshotToSave.Metadata.Index)

	if err := rs.cluster.Recover(&snapshotToSave); err != nil {
		return err
	}

	rs.recoverTransport()

	return updateProgress()
}

func (rs *raftServer) recoverTransport() {
	logger.Info().Msg("remove all peers")
	rs.transport.RemoveAllPeers()

	for _, m := range rs.cluster.AppliedMembers().MapByID {
		if m.ID == rs.cluster.NodeID() {
			continue
		}

		logger.Info().Str("member", m.ToString()).Msg("add raft peer")
		rs.transport.AddPeer(etcdtypes.ID(uint64(m.ID)), []string{m.Url})
	}
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

func unmarshalConfChangeEntry(entry *raftpb.Entry) (*raftpb.ConfChange, *consensus.Member, error) {
	var cc raftpb.ConfChange

	if err := cc.Unmarshal(entry.Data); err != nil {
		logger.Fatal().Err(err).Uint64("idx", entry.Index).Uint64("term", entry.Term).Msg("failed to unmarshal of conf change entry")
		return nil, nil, err
	}

	// skip confChange of empty context
	if len(cc.Context) == 0 {
		return nil, nil, nil
	}

	var member = consensus.Member{}
	if err := json.Unmarshal(cc.Context, &member); err != nil {
		logger.Fatal().Err(err).Uint64("idx", entry.Index).Uint64("term", entry.Term).Msg("failed to unmarshal of context of cc entry")
		return nil, nil, err
	}

	return &cc, &member, nil
}

func (rs *raftServer) ValidateConfChangeEntry(entry *raftpb.Entry) (*raftpb.ConfChange, *consensus.Member, error) {
	// TODO XXX validate from current cluster configure
	var cc *raftpb.ConfChange
	var member *consensus.Member
	var err error

	alreadyApplied := func(entry *raftpb.Entry) bool {
		return rs.cluster.appliedTerm >= entry.Term || rs.cluster.appliedIndex >= entry.Index
	}

	cc, member, err = unmarshalConfChangeEntry(entry)
	if err != nil {
		logger.Fatal().Err(err).Str("entry", entry.String()).Uint64("requestID", cc.ID).Msg("failed to unmarshal conf change")
	}

	if alreadyApplied(entry) {
		return cc, member, ErrCCAlreadyApplied
	}

	if err = rs.cluster.validateChangeMembership(cc, member, true); err != nil {
		return cc, member, err
	}

	return cc, member, nil
}

// TODO refactoring by cc.Type
//      separate unmarshal & apply[type]
// applyConfChange returns false if this node is removed from cluster
func (rs *raftServer) applyConfChange(ent *raftpb.Entry) bool {
	var cc *raftpb.ConfChange
	var member *consensus.Member
	var err error

	postWork := func(err error) bool {
		if err != nil {
			cc.NodeID = raftlib.None
			rs.node.ApplyConfChange(*cc)
		}

		if cc.ID != 0 {
			rs.saveConfChangeState(cc.ID, types.ConfChangeState_CONF_CHANGE_STATE_APPLIED, err)
			rs.cluster.AfterConfChange(cc, member, err)
		}
		return true
	}

	// ConfChanges may be applied more than once. This is because cluster information is more up-to-date than block information when a snapshot is received.
	if cc, member, err = rs.ValidateConfChangeEntry(ent); err != nil {
		logger.Warn().Err(err).Str("entry", types.RaftEntryToString(ent)).Str("cluster", rs.cluster.toString()).Msg("failed to validate conf change")
		return postWork(err)
	}

	rs.confState = rs.node.ApplyConfChange(*cc)

	logger.Info().Uint64("requestID", cc.ID).Str("type", cc.Type.String()).Str("member", member.ToString()).Msg("publish conf change entry")

	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		if err := rs.cluster.addMember(member, true); err != nil {
			logger.Fatal().Str("member", member.ToString()).Msg("failed to add member to cluster")
		}

		if len(cc.Context) > 0 && rs.ID() != cc.NodeID {
			rs.transport.AddPeer(etcdtypes.ID(cc.NodeID), []string{member.Url})
		} else {
			logger.Debug().Msg("skip add peer myself for addnode ")
		}
	case raftpb.ConfChangeRemoveNode:
		if err := rs.cluster.removeMember(member); err != nil {
			logger.Fatal().Str("member", member.ToString()).Msg("failed to add member to cluster")
		}

		if cc.NodeID == rs.ID() {
			logger.Info().Msg("I've been removed from the cluster! Shutting down.")
			return false
		}
		rs.transport.RemovePeer(etcdtypes.ID(cc.NodeID))
	}

	logger.Debug().Uint64("requestID", cc.ID).Str("cluster", rs.cluster.toString()).Msg("after conf changed")

	return postWork(nil)
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rs *raftServer) publishEntries(ents []raftpb.Entry) bool {
	var lastBlockEnt *raftpb.Entry

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

				if block != nil {
					logger.Info().Str("hash", block.ID()).Uint64("no", block.BlockNo()).Msg("commit normal block entry")
					rs.commitProgress.UpdateRequest(&commitEntry{block: block, index: ents[i].Index, term: ents[i].Term})
				}
			}

			select {
			case rs.commitC <- &commitEntry{block: block, index: ents[i].Index, term: ents[i].Term}:
			case <-rs.stopc:
				return false
			}

		case raftpb.EntryConfChange:
			if !rs.applyConfChange(&ents[i]) {
				return false
			}
		}

		// after commit, update appliedIndex
		rs.setAppliedIndex(ents[i].Index)
	}

	if lastBlockEnt != nil {
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

func (rs *raftServer) setConfState(state *raftpb.ConfState) {
	logger.Debug().Str("state", consensus.ConfStateToString(state)).Msg("raft server set confstate")

	rs.confState = state
}

func (rs *raftServer) Process(ctx context.Context, m raftpb.Message) error {
	return rs.node.Step(ctx, m)
}

func (rs *raftServer) IsIDRemoved(id uint64) bool {
	return rs.cluster.IsIDRemoved(id)
}

func (rs *raftServer) ReportUnreachable(id uint64) {
	logger.Debug().Str("toID", EtcdIDToString(id)).Msg("report unreachable")

	rs.node.ReportUnreachable(id)
}

func (rs *raftServer) ReportSnapshot(id uint64, status raftlib.SnapshotStatus) {
	logger.Info().Str("toID", EtcdIDToString(id)).Bool("isSucceed", status == raftlib.SnapshotFinish).Msg("finished to send snapshot")

	rs.node.ReportSnapshot(id, status)
}

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

		logger.Info().Str("ID", EtcdIDToString(rs.ID())).Str("leader", EtcdIDToString(softState.Lead)).Msg("leader changed")
	} else {
		logger.Info().Str("ID", EtcdIDToString(rs.ID())).Str("leader", EtcdIDToString(softState.Lead)).Msg("soft state leader unchanged")
	}
}

func (rs *raftServer) GetLeader() uint64 {
	return atomic.LoadUint64(&rs.leaderStatus.leader)
}

func (rs *raftServer) IsLeader() bool {
	return rs.ID() != consensus.InvalidMemberID && rs.ID() == rs.GetLeader()
}

func (rs *raftServer) Status() raftlib.Status {
	node := rs.getNodeSync()
	if node == nil {
		return raftlib.Status{}
	}

	return node.Status()
}

type MemberProgressState int32

const (
	MemberProgressStateHealthy MemberProgressState = iota
	MemberProgressStateSlow
	MemberProgressStateSyncing
	MemberProgressStateUnknown
)

var (
	MemberProgressStateNames = map[MemberProgressState]string{
		MemberProgressStateHealthy: "MemberProgressStateHealthy",
		MemberProgressStateSlow:    "MemberProgressStateSlow",
		MemberProgressStateSyncing: "MemberProgressStateSyncing",
		MemberProgressStateUnknown: "MemberProgressStateUnknown",
	}
)

type MemberProgress struct {
	MemberID      uint64
	Status        MemberProgressState
	LogDifference uint64

	progress raftlib.Progress
}

type ClusterProgress struct {
	N int

	MemberProgresses map[uint64]*MemberProgress
}

func (cp *ClusterProgress) ToString() string {
	buf := fmt.Sprintf("{ Total: %d, Members[", cp.N)

	for _, mp := range cp.MemberProgresses {
		buf = buf + mp.ToString()
	}

	buf = buf + fmt.Sprintf("] }")

	return buf
}

func (cp *MemberProgress) ToString() string {
	return fmt.Sprintf("{ id: %x, Staus: \"%s\", LogDifference: %d }", cp.MemberID, MemberProgressStateNames[cp.Status], cp.LogDifference)
}

func (rs *raftServer) GetLastIndex() (uint64, error) {
	return rs.raftStorage.LastIndex()
}

func (rs *raftServer) GetClusterProgress() (*ClusterProgress, error) {
	getProgressState := func(raftProgress *raftlib.Progress, lastLeaderIndex uint64, nodeID uint64, leadID uint64) MemberProgressState {
		isLeader := nodeID == leadID

		if !isLeader {
			// syncing
			if raftProgress.State == raftlib.ProgressStateSnapshot {
				return MemberProgressStateSyncing
			}

			// slow
			// - Even if node state is ProgressStateReplicate, if matchNo of the node is too lower than last index of leader, it is considered a slow state.
			var isSlowFollower bool
			if lastLeaderIndex > raftProgress.Match && (lastLeaderIndex-raftProgress.Match) > MaxProgressDiff {
				isSlowFollower = true
			}

			if raftProgress.State == raftlib.ProgressStateProbe || isSlowFollower {
				return MemberProgressStateSlow
			}
		}
		// normal
		return MemberProgressStateHealthy
	}

	var (
		lastIdx uint64
		err     error
	)

	prog := ClusterProgress{}

	node := rs.getNodeSync()
	if node == nil || !rs.IsLeader() {
		return &prog, nil
	}

	status := node.Status()

	n := len(status.Progress)
	if n == 0 {
		return &prog, nil
	}

	statusBytes, err := status.MarshalJSON()
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshalEntryData raft status")
	} else {
		logger.Debug().Str("status", string(statusBytes)).Msg("raft status")
	}

	if lastIdx, err = rs.GetLastIndex(); err != nil {
		logger.Error().Err(err).Msg("Get last raft index on leader")
		return &prog, err
	}

	prog.MemberProgresses = make(map[uint64]*MemberProgress)
	prog.N = n
	for id, nodeProgress := range status.Progress {
		prog.MemberProgresses[id] = &MemberProgress{MemberID: id, Status: getProgressState(&nodeProgress, lastIdx, rs.cluster.NodeID(), id), LogDifference: lastIdx - nodeProgress.Match, progress: nodeProgress}
	}

	return &prog, nil
}

// GetExistingCluster returns information of existing cluster.
// It request member info to all of peers.
func (rs *raftServer) GetExistingCluster() (*Cluster, *types.HardStateInfo, error) {
	var (
		cl        *Cluster
		hardstate *types.HardStateInfo
		err       error
		bestHash  []byte
		bestBlk   *types.Block
	)

	getBestHash := func() []byte {
		if bestBlk, err = rs.walDB.GetBestBlock(); err != nil {
			logger.Fatal().Msg("failed to get best block of my chain to get existing cluster info")
		}

		logger.Info().Str("hash", bestBlk.ID()).Uint64("no", bestBlk.BlockNo()).Msg("best block of backup")

		if bestBlk.BlockNo() == 0 {
			return nil
		}

		return bestBlk.BlockHash()
	}

	bestHash = getBestHash()

	for i := 1; i <= MaxTryGetCluster; i++ {
		cl, hardstate, err = GetClusterInfo(rs.ComponentHub, bestHash)
		if err != nil {
			if err != ErrGetClusterTimeout && i != MaxTryGetCluster {
				logger.Debug().Err(err).Int("try", i).Msg("failed try to get cluster. and sleep")
				time.Sleep(time.Second * 10)
			} else {
				logger.Warn().Err(err).Int("try", i).Msg("failed try to get cluster")
			}
			continue
		}

		return cl, hardstate, nil
	}

	return nil, nil, ErrGetClusterFail
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
