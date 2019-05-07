package raftv2

import (
	"errors"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/consensus/chain"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/etcd/raft/raftpb"
	"github.com/libp2p/go-libp2p-peer"
	"io"
	"time"
)

var (
	DfltTimeWaitPeerLive        = time.Second * 5
	ErrNotMsgSnap               = errors.New("not pb.MsgSnap")
	ErrClusterMismatchConfState = errors.New("members of cluster doesn't match with raft confstate")
)

type getLeaderFuncType func() uint64

type ChainSnapshotter struct {
	p2pcommon.PeerAccessor

	*component.ComponentHub
	cluster *Cluster

	walDB *WalDB

	getLeaderFunc getLeaderFuncType
}

func newChainSnapshotter(pa p2pcommon.PeerAccessor, hub *component.ComponentHub, cluster *Cluster, walDB *WalDB, getLeader getLeaderFuncType) *ChainSnapshotter {
	return &ChainSnapshotter{PeerAccessor: pa, ComponentHub: hub, cluster: cluster, walDB: walDB, getLeaderFunc: getLeader}
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

	if !cluster.isMatch(confstate) {
		logger.Error().Str("confstate", consensus.ConfStateToString(confstate)).Str("cluster", cluster.toString()).Msg("cluster doesn't match with confstate")
		return nil, ErrClusterMismatchConfState
	}

	if cluster.effectiveMembers != cluster.members {
		return nil, ErrNotExistRuntimeMembers
	}
	members := cluster.effectiveMembers.ToArray()

	snap := consensus.NewSnapshotData(members, snapBlock)
	if snap == nil {
		panic("new snap failed")
	}

	return snap, nil
}

// chainSnapshotter rece ives snapshot from http request
// TODO replace rafthttp with p2p
func (chainsnap *ChainSnapshotter) SaveFromRemote(r io.Reader, id uint64, msg raftpb.Message) (int64, error) {
	if msg.Type != raftpb.MsgSnap {
		logger.Error().Int32("type", int32(msg.Type)).Msg("received msg snap is invalid type")
		return 0, ErrNotMsgSnap
	}

	//receive chain & request sync & wait
	return 0, chainsnap.syncSnap(&msg.Snapshot)
}

func (chainsnap *ChainSnapshotter) syncSnap(snap *raftpb.Snapshot) error {
	var snapdata = &consensus.SnapshotData{}
	err := snapdata.Decode(snap.Data)
	if err != nil {
		logger.Fatal().Msg("failed to unmarshal snapshot data to write")
		return err
	}

	// write snapshot log in WAL for crash recovery
	logger.Info().Str("snap", consensus.SnapToString(snap, snapdata)).Msg("start to sync snapshot")
	// TODO	request sync for chain with snapshot.data
	// wait to finish sync of chain
	if err := chainsnap.requestSync(&snapdata.Chain); err != nil {
		logger.Fatal().Err(err).Msg("failed to sync. need to retry with other leader, try N times and shutdown")
		return err
	}

	logger.Info().Str("snap", consensus.SnapToString(snap, snapdata)).Msg("finished to sync snapshot")

	return nil
}

// TODO handle error case that leader stops while synchronizing
func (chainsnap *ChainSnapshotter) requestSync(snap *consensus.ChainSnapshot) error {
	checkPeerLive := func(peerID peer.ID) bool {
		_, ok := chainsnap.GetPeer(peerID)
		return ok
	}

	var leader uint64
	getSyncLeader := func() (peer.ID, error) {
		var peerID peer.ID
		var err error

		for {
			leader = chainsnap.getLeaderFunc()

			if leader == HasNoLeader {
				peerID, err = chainsnap.cluster.getAnyPeerAddressToSync()
				if err != nil {
					logger.Error().Err(err).Str("leader", MemberIDToString(leader)).Msg("can't get peeraddress of leader")
					return "", err
				}
			} else {
				peerID, err = chainsnap.cluster.getEffectiveMembers().getMemberPeerAddress(leader)
				if err != nil {
					logger.Error().Err(err).Str("leader", MemberIDToString(leader)).Msg("can't get peeraddress of leader")
					return "", err
				}
			}

			if checkPeerLive(peerID) {
				break
			}

			logger.Debug().Str("peer", p2putil.ShortForm(peerID)).Str("leader", MemberIDToString(leader)).Msg("peer is not alive")

			time.Sleep(DfltTimeWaitPeerLive)
		}

		logger.Debug().Str("peer", p2putil.ShortForm(peerID)).Str("leader", MemberIDToString(leader)).Msg("target peer to sync")

		return peerID, err
	}

	peerID, err := getSyncLeader()
	if err != nil {
		return err
	}

	if err := chain.SyncChain(chainsnap.ComponentHub, snap.Hash, snap.No, peerID); err != nil {
		return err
	}

	return nil
}
