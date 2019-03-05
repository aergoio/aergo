package raft

import (
	"errors"
	"github.com/aergoio/aergo/config"
	"net"
	"net/url"
	"os"
	"strings"
)

var (
	ErrInvalidRaftID    = errors.New("invalid raft id")
	ErrDupRaftUrl       = errors.New("duplicated raft bp urls")
	ErrRaftEmptyTLSFile = errors.New("cert or key file name is empty")
	ErrNotHttpsURL      = errors.New("url scheme is not https")
	ErrURLInvalidScheme = errors.New("url has invalid scheme")
	ErrURLInvalidPort   = errors.New("url must have host:port style")
)

func checkConfig(cfg *config.Config) error {
	useTls := true
	var err error

	if useTls, err = validateTLS(cfg.Consensus); err != nil {
		logger.Error().Err(err).
			Str("key", cfg.Consensus.RaftKeyFile).
			Str("cert", cfg.Consensus.RaftCertFile).
			Msg("failed to validate tls config for raft")
		return err
	}

	var outUrls []string
	if outUrls, err = validateBPUrls(cfg.Consensus, useTls); err != nil {
		logger.Error().Err(err).Msg("failed to validate bpurls, bpid config for raft")
		return err
	}

	cfg.Consensus.RaftBpUrls = outUrls

	RaftSkipEmptyBlock = cfg.Consensus.RaftSkipEmpty
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

func validateBPUrls(consCfg *config.ConsensusConfig, useTls bool) ([]string, error) {
	//TODO check Url
	// - each urlstr is unique
	// - format is valid urlstr
	//check ID
	// - 1 <= ID <= len(RaftBpUrls)

	lenBpUrls := len(consCfg.RaftBpUrls)
	outUrls := make([]string, lenBpUrls)

	var urlobj *url.URL
	var err error

	urlMap := make(map[string]bool, lenBpUrls)
	for i, urlstr := range consCfg.RaftBpUrls {
		urlstr = strings.TrimSpace(urlstr)
		if _, ok := urlMap[urlstr]; ok {
			return nil, ErrDupRaftUrl
		} else {
			urlMap[urlstr] = true
		}

		if urlobj, err = parseToUrl(urlstr); err != nil {
			logger.Error().Str("url", urlstr).Err(err).Msg("raft bp urlstr is not vaild form")
			return nil, err
		}

		if useTls && urlobj.Scheme != "https" {
			logger.Error().Str("urlstr", urlstr).Msg("raft bp urlstr shoud use https protocol")
			return nil, ErrNotHttpsURL
		}

		outUrls[i] = urlstr
	}

	raftID := consCfg.RaftID
	if raftID <= 0 || raftID > uint64(lenBpUrls) {
		return nil, ErrInvalidRaftID
	}

	return outUrls, nil
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
