package raftv2

import (
	"errors"
	"time"

	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
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
func GetClusterInfo(hs *component.ComponentHub, bestHash []byte) (*Cluster, *types.HardStateInfo, error) {
	logger.Info().Msg("try getclusterinfo to p2p")

	replyC := make(chan *message.GetClusterRsp)
	hs.Tell(message.P2PSvc, &message.GetCluster{BestBlockHash: bestHash, ReplyC: replyC})

	var (
		rsp   *message.GetClusterRsp
		ok    bool
		err   error
		newCl *Cluster
	)

	select {
	case rsp, ok = <-replyC:
		if !ok {
			return nil, nil, ErrGetClusterReplyC
		}

		if rsp.Err != nil {
			return nil, nil, rsp.Err
		}

		if len(rsp.Members) == 0 {
			return nil, nil, ErrGetClusterEmpty
		}

	case <-time.After(MaxTimeOutCluter):
		return nil, nil, ErrGetClusterTimeout
	}

	if newCl, err = NewClusterFromMemberAttrs(rsp.ClusterID, rsp.ChainID, rsp.Members); err != nil {
		return nil, nil, err
	}

	//logger.Debug().Str("info", newCl.toString()).Msg("get remote cluster info")
	return newCl, rsp.HardStateInfo, nil
}
