package raftv2

import (
	"errors"
	"io"
	"sync"
	"time"

	chainsvc "github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/consensus/chain"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/etcd/raft/raftpb"
)

var (
	DfltTimeWaitPeerLive        = time.Second * 5
	ErrNotMsgSnap               = errors.New("not pb.MsgSnap")
	ErrClusterMismatchConfState = errors.New("members of cluster doesn't match with raft confstate")
)

type getLeaderFuncType func() uint64

type ChainSnapshotter struct {
	sync.Mutex

	pa p2pcommon.PeerAccessor

	*component.ComponentHub
	cluster *Cluster

	walDB *WalDB

	getLeaderFunc getLeaderFuncType
}

func newChainSnapshotter(pa p2pcommon.PeerAccessor, hub *component.ComponentHub, cluster *Cluster, walDB *WalDB, getLeader getLeaderFuncType) *ChainSnapshotter {
	return &ChainSnapshotter{pa: pa, ComponentHub: hub, cluster: cluster, walDB: walDB, getLeaderFunc: getLeader}
}

func (chainsnap *ChainSnapshotter) setPeerAccessor(pa p2pcommon.PeerAccessor) {
	chainsnap.Lock()
	defer chainsnap.Unlock()

	chainsnap.pa = pa
}

/* createSnapshot isn't used this api since new MsgSnap isn't made
// createSnapshot make marshalled data of chain & cluster info
func (chainsnap *ChainSnapshotter) createSnapshot(prevProgress BlockProgress, confState raftpb.ConfState) (*raftpb.Snapshot, error) {
	if prevProgress.isEmpty() {
		return nil, ErrEmptyProgress
	}

	snapdata, err := chainsnap.createSnapshotData(chainsnap.cluster, prevProgress.block)
	if err != nil {
		logger.Fatal().Err(err).Msg("make snapshot of chain")
		return nil, err
	}


	data, err := snapdata.Encode()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to marshale snapshot of chain")
		return nil, err
	}

	snapshot := &raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{
			Index:     prevProgress.index,
			Term:      prevProgress.term,
			ConfState: confState,
		},
		Data: data,
	}

	logger.Info().Str("snapshot", consensus.SnapToString(snapshot, snapdata)).Msg("raft snapshot for remote")

	return snapshot, nil
}
*/

// createSnapshotData generate serialized data of chain and cluster info
func (chainsnap *ChainSnapshotter) createSnapshotData(cluster *Cluster, snapBlock *types.Block, confstate *raftpb.ConfState) (*consensus.SnapshotData, error) {
	logger.Info().Str("hash", snapBlock.ID()).Uint64("no", snapBlock.BlockNo()).Msg("create new snapshot data of block")

	cluster.Lock()
	defer cluster.Unlock()

	if !cluster.isMatch(confstate) {
		logger.Fatal().Str("confstate", consensus.ConfStateToString(confstate)).Str("cluster", cluster.toStringWithLock()).Msg("cluster doesn't match with confstate")
		return nil, ErrClusterMismatchConfState
	}

	members := cluster.AppliedMembers().ToArray()
	removedMembers := cluster.RemovedMembers().ToArray()

	snap := consensus.NewSnapshotData(members, removedMembers, snapBlock)
	if snap == nil {
		logger.Panic().Msg("new snap failed")
	}

	return snap, nil
}

// chainSnapshotter rece ives snapshot from http request
// TODO replace rafthttp with p2p
func (chainsnap *ChainSnapshotter) SaveFromRemote(r io.Reader, id uint64, msg raftpb.Message) (int64, error) {
	defer RecoverExit()

	if msg.Type != raftpb.MsgSnap {
		logger.Error().Int32("type", int32(msg.Type)).Msg("received msg snap is invalid type")
		return 0, ErrNotMsgSnap
	}

	// not return until block sync is complete
	// receive chain & request sync & wait
	return 0, chainsnap.syncSnap(&msg.Snapshot)
}

func (chainsnap *ChainSnapshotter) syncSnap(snap *raftpb.Snapshot) error {
	var snapdata = &consensus.SnapshotData{}

	err := snapdata.Decode(snap.Data)
	if err != nil {
		logger.Error().Msg("failed to unmarshal snapshot data to write")
		return err
	}

	// write snapshot log in WAL for crash recovery
	logger.Info().Str("snap", consensus.SnapToString(snap, snapdata)).Msg("start to sync snapshot")
	// TODO	request sync for chain with snapshot.data
	// wait to finish sync of chain
	if err := chainsnap.requestSync(&snapdata.Chain); err != nil {
		logger.Error().Err(err).Msg("failed to sync snapshot")
		return err
	}

	logger.Info().Str("snap", consensus.SnapToString(snap, snapdata)).Msg("finished to sync snapshot")

	return nil
}

func (chainsnap *ChainSnapshotter) checkPeerLive(peerID types.PeerID) bool {
	if chainsnap.pa == nil {
		logger.Fatal().Msg("peer accessor of chain snapshotter is not set")
	}

	_, ok := chainsnap.pa.GetPeer(peerID)
	return ok
}

// TODO handle error case that leader stops while synchronizing
func (chainsnap *ChainSnapshotter) requestSync(snap *consensus.ChainSnapshot) error {

	var leader uint64
	getSyncLeader := func() (types.PeerID, error) {
		var peerID types.PeerID
		var err error

		for {
			leader = chainsnap.getLeaderFunc()

			if leader == HasNoLeader {
				peerID, err = chainsnap.cluster.getAnyPeerAddressToSync()
				if err != nil {
					logger.Error().Err(err).Str("leader", EtcdIDToString(leader)).Msg("can't get peeraddress of leader")
					return "", err
				}
			} else {
				peerID, err = chainsnap.cluster.Members().getMemberPeerAddress(leader)
				if err != nil {
					logger.Error().Err(err).Str("leader", EtcdIDToString(leader)).Msg("can't get peeraddress of leader")
					return "", err
				}
			}

			if chainsnap.checkPeerLive(peerID) {
				break
			}

			logger.Debug().Stringer("peer", types.LogPeerShort(peerID)).Str("leader", EtcdIDToString(leader)).Msg("peer is not alive")

			time.Sleep(DfltTimeWaitPeerLive)
		}

		logger.Debug().Stringer("peer", types.LogPeerShort(peerID)).Str("leader", EtcdIDToString(leader)).Msg("target peer to sync")

		return peerID, err
	}

	chainsvc.TestDebugger.Check(chainsvc.DEBUG_SYNCER_CRASH, 1, nil)

	peerID, err := getSyncLeader()
	if err != nil {
		return err
	}

	if err := chain.SyncChain(chainsnap.ComponentHub, snap.Hash, snap.No, peerID); err != nil {
		return err
	}

	return nil
}
