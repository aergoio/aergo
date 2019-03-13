package raft

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo/config"
	"github.com/libp2p/go-libp2p-peer"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	RaftTick = DefaultTickMS

	ErrInvalidRaftID     = errors.New("invalid raft raftID")
	ErrDupRaftUrl        = errors.New("duplicated raft bp urls")
	ErrRaftEmptyTLSFile  = errors.New("cert or key file name is empty")
	ErrNotHttpsURL       = errors.New("url scheme is not https")
	ErrURLInvalidScheme  = errors.New("url has invalid scheme")
	ErrURLInvalidPort    = errors.New("url must have host:port style")
	ErrInvalidRaftBPID   = errors.New("raft bp raftID is not ordered. raftID must start with 1 and be sorted")
	ErrDupBP             = errors.New("raft bp description is duplicated")
	ErrInvalidRaftPeerID = errors.New("peerID of current raft bp is not equals to p2p configure")
)

const (
	DefaultMarginChainDiff = 1
	DefaultTickMS          = time.Millisecond * 30
)

func (bf *BlockFactory) InitCluster(cfg *config.Config) error {
	useTls := true
	var err error

	raftConfig := cfg.Consensus.Raft
	if raftConfig == nil {
		panic("raftconfig is not set. please set raftID, raftBPs.")
	}

	//set default
	if raftConfig.RaftTick != 0 {
		RaftTick = RaftTick
	}

	lenBPs := len(raftConfig.RaftBPs)
	raftID := raftConfig.RaftID

	bf.bpc = NewCluster(bf, raftID, uint16(lenBPs))

	if raftID <= 0 || raftID > uint64(lenBPs) {
		logger.Error().Err(err).Msg("raft raftID has the following values: 1 <= raft raftID <= len(bpcount)")

		return ErrInvalidRaftID
	}

	if useTls, err = validateTLS(raftConfig); err != nil {
		logger.Error().Err(err).
			Str("key", raftConfig.RaftKeyFile).
			Str("cert", raftConfig.RaftCertFile).
			Msg("failed to validate tls config for raft")
		return err
	}

	if err = bf.bpc.addMembers(raftConfig, useTls); err != nil {
		logger.Error().Err(err).Msg("failed to validate bpurls, bpid config for raft")
		return err
	}

	RaftSkipEmptyBlock = raftConfig.RaftSkipEmpty

	logger.Info().Msg(bf.bpc.toString())

	return nil
}

func validateTLS(raftCfg *config.RaftConfig) (bool, error) {
	if len(raftCfg.RaftCertFile) == 0 && len(raftCfg.RaftKeyFile) == 0 {
		return false, nil
	}

	//두 파일이 모두 설정되어 있는지 확인
	//실제 file에 존재하는지 확인
	if len(raftCfg.RaftCertFile) == 0 || len(raftCfg.RaftKeyFile) == 0 {
		logger.Error().Str("raftcertfile", raftCfg.RaftCertFile).Str("raftkeyfile", raftCfg.RaftKeyFile).
			Msg(ErrRaftEmptyTLSFile.Error())
		return false, ErrRaftEmptyTLSFile
	}

	if len(raftCfg.RaftCertFile) != 0 {
		if _, err := os.Stat(raftCfg.RaftCertFile); err != nil {
			logger.Error().Err(err).Msg("not exist certificate file for raft")
			return false, err
		}
	}

	if len(raftCfg.RaftKeyFile) != 0 {
		if _, err := os.Stat(raftCfg.RaftKeyFile); err != nil {
			logger.Error().Err(err).Msg("not exist Key file for raft")
			return false, err
		}
	}

	return true, nil
}

func isValidURL(urlstr string, useTls bool) error {
	var urlobj *url.URL
	var err error

	if urlobj, err = parseToUrl(urlstr); err != nil {
		logger.Error().Str("url", urlstr).Err(err).Msg("raft bp urlstr is not vaild form")
		return err
	}

	if useTls && urlobj.Scheme != "https" {
		logger.Error().Str("urlstr", urlstr).Msg("raft bp urlstr shoud use https protocol")
		return ErrNotHttpsURL
	}

	return nil
}

func isValidID(raftID uint64, lenBps int) error {
	if raftID <= 0 || raftID > uint64(lenBps) {
		logger.Error().Msg("raft raftID has the following values: 1 <= raft raftID <= len(bpcount)")

		return ErrInvalidRaftID
	}

	return nil
}

func (cc *Cluster) addMembers(raftCfg *config.RaftConfig, useTls bool) error {
	lenBPs := len(raftCfg.RaftBPs)
	if lenBPs == 0 {
		return fmt.Errorf("config of raft bp is empty")
	}

	// validate each bp
	for i, raftBP := range raftCfg.RaftBPs {
		if uint64(i+1) != raftBP.ID {
			return ErrInvalidRaftBPID
		}

		urlstr := raftBP.Url
		trimUrl := strings.TrimSpace(urlstr)

		if err := isValidURL(urlstr, useTls); err != nil {
			return err
		}

		if err := isValidID(raftBP.ID, lenBPs); err != nil {
			return err
		}

		peerID, err := peer.IDB58Decode(raftBP.P2pID)
		if err != nil {
			return fmt.Errorf("invalid raft peerID %s", raftBP.P2pID)
		}

		if err := cc.addMember(raftBP.ID, trimUrl, peerID); err != nil {
			return err
		}
	}

	// TODO check my node pubkey from p2p

	return nil
}

func parseToUrl(urlstr string) (*url.URL, error) {
	var urlObj *url.URL
	var err error

	if urlObj, err = url.Parse(urlstr); err != nil {
		return nil, err
	}

	if urlObj.Scheme != "http" && urlObj.Scheme != "https" {
		return nil, ErrURLInvalidScheme
	}

	if _, _, err := net.SplitHostPort(urlObj.Host); err != nil {
		return nil, ErrURLInvalidPort
	}

	return urlObj, nil
}
