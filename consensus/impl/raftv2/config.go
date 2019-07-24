package raftv2

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"os"
	"strings"
	"time"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/consensus"
)

var (
	ErrEmptyBPs              = errors.New("BP list is empty")
	ErrNotIncludedRaftMember = errors.New("this node isn't included in initial raft members")
	ErrRaftEmptyTLSFile      = errors.New("cert or key file name is empty")
	ErrNotHttpsURL           = errors.New("url scheme is not https")
	ErrDupBP                 = errors.New("raft bp description is duplicated")
	ErrInvalidRaftPeerID     = errors.New("peerID of current raft bp is not equals to p2p configure")
	ErrNotExitRecoverBP      = errors.New("RecoverBP is needed for creating a new cluster from backup")
)

const (
	DefaultTickMS = time.Millisecond * 30
)

func (bf *BlockFactory) InitCluster(cfg *config.Config) error {
	useTls := true
	var err error

	genesis := chain.Genesis

	raftConfig := cfg.Consensus.Raft
	if raftConfig == nil {
		panic("raftconfig is not set. please set raftName, raftBPs.")
	}

	//set default
	if raftConfig.Tick != 0 {
		RaftTick = time.Duration(raftConfig.Tick * 1000000)
	}

	if raftConfig.SnapFrequency != 0 {
		ConfSnapFrequency = raftConfig.SnapFrequency
		ConfSnapshotCatchUpEntriesN = raftConfig.SnapFrequency
	}

	chainID, err := genesis.ID.Bytes()
	if err != nil {
		return err
	}

	bf.bpc = NewCluster(chainID, bf, raftConfig.Name, genesis.Timestamp, func(event *message.RaftClusterEvent) { bf.Tell(message.P2PSvc, event) })

	if useTls, err = validateTLS(raftConfig); err != nil {
		logger.Error().Err(err).
			Str("key", raftConfig.KeyFile).
			Str("cert", raftConfig.CertFile).
			Bool("useTLS", useTls).
			Msg("failed to validate tls config for raft")
		return err
	}

	//if raftConfig.ListenUrl != "" {
	//	if err := isValidURL(raftConfig.ListenUrl, useTls); err != nil {
	//		logger.Error().Err(err).Msg("failed to validate listen url for raft")
	//		return err
	//	}
	//}

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
			logger.Error().Err(err).Msg("failed to validate bpurls, bpid config for raft")
			return err
		}

		if bf.bpc.Members().len() == 0 {
			logger.Fatal().Str("cluster", bf.bpc.toString()).Msg("can't start raft server because there are no members in cluster")
		}

		if bf.bpc.Members().getMemberByName(raftConfig.Name) == nil {
			logger.Fatal().Str("cluster", bf.bpc.toString()).Msg("node name of config is not included in genesis block")
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
		if _, err := types.ParseMultiaddrWithResolve(trimmedAddr); err != nil {
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

func validateTLS(raftCfg *config.RaftConfig) (bool, error) {
	if len(raftCfg.CertFile) == 0 && len(raftCfg.KeyFile) == 0 {
		return false, nil
	}

	//두 파일이 모두 설정되어 있는지 확인
	//실제 file에 존재하는지 확인
	if len(raftCfg.CertFile) == 0 || len(raftCfg.KeyFile) == 0 {
		logger.Error().Str("raftcertfile", raftCfg.CertFile).Str("raftkeyfile", raftCfg.KeyFile).
			Msg(ErrRaftEmptyTLSFile.Error())
		return false, ErrRaftEmptyTLSFile
	}

	if len(raftCfg.CertFile) != 0 {
		if _, err := os.Stat(raftCfg.CertFile); err != nil {
			logger.Error().Err(err).Msg("not exist certificate file for raft")
			return false, err
		}
	}

	if len(raftCfg.KeyFile) != 0 {
		if _, err := os.Stat(raftCfg.KeyFile); err != nil {
			logger.Error().Err(err).Msg("not exist Key file for raft")
			return false, err
		}
	}

	return true, nil
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
