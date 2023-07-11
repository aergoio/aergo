package raftv2

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/types"
)

var (
	ErrEmptyBPs              = errors.New("BP list is empty")
	ErrNotIncludedRaftMember = errors.New("this node isn't included in initial raft members")
	ErrDupBP                 = errors.New("raft bp description is duplicated")
	ErrInvalidRaftPeerID     = errors.New("peerID of current raft bp is not equals to p2p configure")
)

const (
	slotQueueMax = 100

	DefaultBlockFactoryTickMs = 100
	MinBlockFactoryTickMs     = 10
	DefaultTickMS             = time.Millisecond * 50
	DefaultElectionTickCount  = 10
	DefaultSlowNodeGap        = 100
	DefaultSnapFrequency      = 30
)

var (
	RaftTick = DefaultTickMS
	// blockIntervalMs is the block genration interval in milli-seconds.
	RaftSkipEmptyBlock = false
	//MaxCommitQueueLen  = DefaultCommitQueueLen

	BlockFactoryTickMs time.Duration
	BlockIntervalMs    time.Duration

	ConfSnapFrequency           uint64 = DefaultSnapFrequency
	ConfSnapshotCatchUpEntriesN uint64 = ConfSnapFrequency

	ElectionTickCount        = DefaultElectionTickCount
	MaxSlowNodeGap    uint64 = DefaultSlowNodeGap // Criteria for determining whether the server is in a slow state
	StopDupCommit            = false
)

func Init(raftCfg *config.RaftConfig) {
	var tickMs time.Duration

	//set default
	if raftCfg.HeartbeatTick != 0 {
		RaftTick = time.Duration(raftCfg.HeartbeatTick * 1000000)
	}

	if raftCfg.ElectionTickCount > 0 {
		ElectionTickCount = raftCfg.ElectionTickCount
	}

	if raftCfg.SnapFrequency != 0 {
		ConfSnapFrequency = raftCfg.SnapFrequency
		ConfSnapshotCatchUpEntriesN = raftCfg.SnapFrequency
	}

	if raftCfg.BlockFactoryTickMs != 0 {
		if raftCfg.BlockFactoryTickMs < MinBlockFactoryTickMs {
			tickMs = MinBlockFactoryTickMs
		} else {
			tickMs = time.Duration(raftCfg.BlockFactoryTickMs)
		}
	} else {
		tickMs = DefaultBlockFactoryTickMs
	}

	BlockFactoryTickMs = time.Millisecond * tickMs

	if raftCfg.BlockIntervalMs != 0 {
		BlockIntervalMs = time.Millisecond * time.Duration(raftCfg.BlockIntervalMs)
	} else {
		BlockIntervalMs = consensus.BlockInterval
	}

	if raftCfg.SlowNodeGap > 0 {
		MaxSlowNodeGap = uint64(raftCfg.SlowNodeGap)
	}

	if raftCfg.StopDupCommit {
		StopDupCommit = true
	}
	logger.Info().Int64("factory tick(ms)", BlockFactoryTickMs.Nanoseconds()/int64(time.Millisecond)).
		Int64("interval(ms)", BlockIntervalMs.Nanoseconds()/int64(time.Millisecond)).Msg("set block factory tick/interval")
}

func (bf *BlockFactory) InitCluster(cfg *config.Config) error {
	var err error

	genesis := chain.Genesis

	raftConfig := cfg.Consensus.Raft
	if raftConfig == nil {
		logger.Panic().Msg("raft config is not set. please set raftName, raftBPs.")
	}

	chainID, err := genesis.ID.Bytes()
	if err != nil {
		return err
	}

	bf.bpc = NewCluster(chainID, bf, raftConfig.Name, p2pkey.NodeID(), genesis.Timestamp, func(event *message.RaftClusterEvent) { bf.Tell(message.P2PSvc, event) })

	if raftConfig.NewCluster {
		var mbrAttrs []*types.MemberAttr
		var ebps []types.EnterpriseBP

		if !raftConfig.UseBackup {
			ebps = chain.Genesis.EnterpriseBPs
		} else {
			ebps = getRecoverBp(raftConfig)
		}

		if mbrAttrs, err = parseBpsToMembers(ebps); err != nil {
			logger.Error().Err(err).Bool("usebackup", raftConfig.UseBackup).Msg("failed to parse initial bp list")
			return err
		}

		if err = bf.bpc.AddInitialMembers(mbrAttrs); err != nil {
			logger.Error().Err(err).Msg("failed to add initial members")
			return err
		}
	}

	RaftSkipEmptyBlock = raftConfig.SkipEmpty

	logger.Info().Bool("skipempty", RaftSkipEmptyBlock).Int64("rafttick(nanosec)", RaftTick.Nanoseconds()).Float64("interval(sec)", consensus.BlockInterval.Seconds()).Msg(bf.bpc.toString())

	return nil
}

// getRecoverBp returns Enterprise BP to use initial bp of new cluster for recovery from backup
func getRecoverBp(raftConfig *config.RaftConfig) []types.EnterpriseBP {
	if raftConfig.RecoverBP == nil {
		logger.Fatal().Msg("need RecoverBP in config to create a new cluster")
	}

	cfgBP := raftConfig.RecoverBP
	return []types.EnterpriseBP{{Name: cfgBP.Name, Address: cfgBP.Address, PeerID: cfgBP.PeerID}}
}

func parseBpsToMembers(bps []types.EnterpriseBP) ([]*types.MemberAttr, error) {
	bpLen := len(bps)
	if bpLen == 0 {
		return nil, ErrEmptyBPs
	}

	mbrs := make([]*types.MemberAttr, bpLen)
	for i, bp := range bps {
		trimmedAddr := strings.TrimSpace(bp.Address)
		// TODO when p2p is applied, have to validate peer address
		if _, err := types.ParseMultiaddr(trimmedAddr); err != nil {
			return nil, err
		}

		peerID, err := types.IDB58Decode(bp.PeerID)
		if err != nil {
			return nil, fmt.Errorf("invalid raft peerID BP[%d]:%s", i, bp.PeerID)
		}

		mbrs[i] = &types.MemberAttr{Name: bp.Name, Address: trimmedAddr, PeerID: []byte(peerID)}
	}

	return mbrs, nil
}

func (cl *Cluster) AddInitialMembers(mbrs []*types.MemberAttr) error {
	logger.Debug().Msg("add cluster members from config file")

	for _, mbrAttr := range mbrs {
		m := consensus.NewMember(mbrAttr.Name, mbrAttr.Address, types.PeerID(mbrAttr.PeerID), cl.chainID, cl.chainTimestamp)

		if err := cl.isValidMember(m); err != nil {
			return err
		}
		if err := cl.addMember(m, false); err != nil {
			return err
		}
	}

	if cl.Members().len() == 0 {
		logger.Fatal().Str("cluster", cl.toString()).Msg("can't start raft server because there are no members in cluster")
	}

	if cl.Members().getMemberByName(cl.NodeName()) == nil {
		logger.Fatal().Str("cluster", cl.toString()).Msg("node name of config is not included in genesis block")
	}

	return nil
}

func (cl *Cluster) SetThisNodeID() error {
	cl.Lock()
	defer cl.Unlock()

	var member *consensus.Member

	if member = cl.Members().getMemberByName(cl.NodeName()); member == nil {
		return ErrNotIncludedRaftMember
	}

	// it can be reset when this node is added to cluster
	cl.SetNodeID(member.ID)

	return nil
}
