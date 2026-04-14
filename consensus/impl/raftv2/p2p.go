package raftv2

import (
	"errors"
	"time"

	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/message"
)

var (
	MaxTimeOutCluter = time.Second * 10
	MaxTryGetCluster = 3

	ErrGetClusterReplyC  = errors.New("reply channel of getcluster request is closed")
	ErrGetClusterTimeout = errors.New("timeout for getcluster")
	ErrGetClusterEmpty   = errors.New("getcluster reply is empty")
	ErrGetClusterFail    = errors.New("failed to get cluster info")
)

// GetClusterInfo collects cluster info from remote peers via p2p module.
// It returns the cluster, the hardstate of the remote best block, the remote best block number, and an error.
func GetClusterInfo(hs *component.ComponentHub, bestHash []byte) (*Cluster, *types.HardStateInfo, types.BlockNo, error) {
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
			return nil, nil, 0, ErrGetClusterReplyC
		}

		if rsp.Err != nil {
			return nil, nil, 0, rsp.Err
		}

		if len(rsp.Members) == 0 {
			return nil, nil, 0, ErrGetClusterEmpty
		}

	case <-time.After(MaxTimeOutCluter):
		return nil, nil, 0, ErrGetClusterTimeout
	}

	if newCl, err = NewClusterFromMemberAttrs(rsp.ClusterID, rsp.ChainID, rsp.Members); err != nil {
		return nil, nil, 0, err
	}

	//logger.Debug().Str("info", newCl.toString()).Msg("get remote cluster info")
	return newCl, rsp.HardStateInfo, rsp.BestBlockNo, nil
}
