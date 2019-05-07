package raftv2

import (
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"time"
)

// GetBestBlock returns the current best block from chainservice
func GetClusterMembers(hs component.ICompSyncRequester) ([]*consensus.Member, error) {
	result, err := hs.RequestFuture(message.P2PSvc, &message.GetClusterMembers{}, time.Second*10,
		"consensus/raft/info.GetClusterMembers").Result()

	if result.(message.GetClusterMembersRsp).Err != nil {
		err = result.(message.GetClusterMembersRsp).Err
	}

	if err != nil {
		logger.Error().Err(err).Msg("failed to get cluster members from remote peers")
		return nil, err
	}

	mbrs := make([]*consensus.Member, len(result.(message.GetClusterMembersRsp).Members))

	for i, mbrAttr := range result.(message.GetClusterMembersRsp).Members {
		mbrs[i].SetAttr(mbrAttr)
	}

	return mbrs, nil
}
