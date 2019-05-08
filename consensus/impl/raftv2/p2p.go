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

	ErrClusterReplyC     = errors.New("reply channel of getcluster request is closed")
	ErrGetClusterTimeout = errors.New("timeout for getcluster")
)

// GetBestBlock returns the current best block from chainservice
func GetClusterInfo(hs *component.ComponentHub) (*Cluster, error) {
	replyC := make(chan *message.GetClusterRsp)
	hs.Tell(message.P2PSvc, &message.GetCluster{ReplyC: replyC})

	var (
		rsp *message.GetClusterRsp
		ok  bool
	)

	select {
	case rsp, ok = <-replyC:
		if !ok {
			return nil, ErrClusterReplyC
		}
	case <-time.After(MaxTimeOutCluter):
		return nil, ErrGetClusterTimeout
	}

	if rsp.Err != nil {
		return nil, fmt.Errorf("get cluster failed: %s", rsp.Err)
	}

	newCl := NewClusterFromMemberAttrs(rsp.ChainID, rsp.Members)

	logger.Debug().Str("info", newCl.toString()).Msg("get remote cluster info")
	return newCl, nil
}
