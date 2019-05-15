package raftv2

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"time"
)

var (
	MaxTimeOutCluter = time.Second * 10
	MaxTryGetCluster = 3

	ErrGetClusterReplyC  = errors.New("reply channel of getcluster request is closed")
	ErrGetClusterTimeout = errors.New("timeout for getcluster")
	ErrGetClusterEmpty   = errors.New("getcluster reply is empty")
	ErrGetClusterFail    = errors.New("failed to get cluster info")
)

// GetBestBlock returns the current best block from chainservice
func GetClusterInfo(hs *component.ComponentHub) (*Cluster, error) {
	logger.Info().Msg("try getclusterinfo to p2p")

	replyC := make(chan *message.GetClusterRsp)
	hs.Tell(message.P2PSvc, &message.GetCluster{ReplyC: replyC})

	var (
		rsp *message.GetClusterRsp
		ok  bool
	)

	select {
	case rsp, ok = <-replyC:
		if !ok {
			return nil, ErrGetClusterReplyC
		}

		if rsp.Err != nil {
			return nil, fmt.Errorf("get cluster failed: %s", rsp.Err)
		}

		if len(rsp.Members) == 0 {
			return nil, ErrGetClusterEmpty
		}

	case <-time.After(MaxTimeOutCluter):
		return nil, ErrGetClusterTimeout
	}

	newCl := NewClusterFromMemberAttrs(rsp.ChainID, rsp.Members)

	//logger.Debug().Str("info", newCl.toString()).Msg("get remote cluster info")
	return newCl, nil
}
