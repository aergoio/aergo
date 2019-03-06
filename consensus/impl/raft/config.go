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
)

var (
	ErrInvalidRaftID    = errors.New("invalid raft raftID")
	ErrDupRaftUrl       = errors.New("duplicated raft bp urls")
	ErrRaftEmptyTLSFile = errors.New("cert or key file name is empty")
	ErrNotHttpsURL      = errors.New("url scheme is not https")
	ErrURLInvalidScheme = errors.New("url has invalid scheme")
	ErrURLInvalidPort   = errors.New("url must have host:port style")
	ErrInvalidRaftBPID  = errors.New("raft bp raftID is not ordered. raftID must start with 1 and be sorted")
	ErrDupBP            = errors.New("raft bp description is duplicated")
)

func checkConfig(cfg *config.Config, cluster *Cluster) error {
	useTls := true
	var err error

	raftID := cfg.Consensus.RaftID
	if raftID <= 0 || raftID > uint64(len(cfg.Consensus.RaftBPs)) {
		logger.Error().Err(err).Msg("raft raftID has the following values: 1 <= raft raftID <= len(bpcount)")

		return ErrInvalidRaftID
	}

	if useTls, err = validateTLS(cfg.Consensus); err != nil {
		logger.Error().Err(err).
			Str("key", cfg.Consensus.RaftKeyFile).
			Str("cert", cfg.Consensus.RaftCertFile).
			Msg("failed to validate tls config for raft")
		return err
	}

	if err = validateBPs(cfg.Consensus, useTls, cluster); err != nil {
		logger.Error().Err(err).Msg("failed to validate bpurls, bpid config for raft")
		return err
	}

	RaftSkipEmptyBlock = cfg.Consensus.RaftSkipEmpty

	logger.Info().Msg(cluster.toString())

	return nil
}

func validateTLS(consCfg *config.ConsensusConfig) (bool, error) {
	if len(consCfg.RaftCertFile) == 0 && len(consCfg.RaftKeyFile) == 0 {
		return false, nil
	}

	//두 파일이 모두 설정되어 있는지 확인
	//실제 file에 존재하는지 확인
	if len(consCfg.RaftCertFile) == 0 || len(consCfg.RaftKeyFile) == 0 {
		logger.Error().Str("raftcertfile", consCfg.RaftCertFile).Str("raftkeyfile", consCfg.RaftKeyFile).
			Msg(ErrRaftEmptyTLSFile.Error())
		return false, ErrRaftEmptyTLSFile
	}

	if len(consCfg.RaftCertFile) != 0 {
		if _, err := os.Stat(consCfg.RaftCertFile); err != nil {
			logger.Error().Err(err).Msg("not exist certificate file for raft")
			return false, err
		}
	}

	if len(consCfg.RaftKeyFile) != 0 {
		if _, err := os.Stat(consCfg.RaftKeyFile); err != nil {
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

func validateBPs(consCfg *config.ConsensusConfig, useTls bool, cl *Cluster) error {
	lenBPs := len(consCfg.RaftBPs)

	cl.ID = consCfg.RaftID
	cl.Size = uint16(lenBPs)
	cl.Member = make(map[uint64]*blockProducer)
	cl.BPUrls = make([]string, lenBPs)

	// validate each bp
	for i, raftBP := range consCfg.RaftBPs {
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

		cl.addMember(raftBP.ID, trimUrl, peerID)
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
