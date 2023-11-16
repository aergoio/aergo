package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvCommitResultList(msg *types.CommitResultList) *InOutCommitResultList {
	c := &InOutCommitResultList{}
	c.Results = make([]*InOutCommitResult, len(msg.Results))
	for i, result := range msg.Results {
		c.Results[i] = ConvCommitResult(result)
	}
	return c
}

type InOutCommitResultList struct {
	Results []*InOutCommitResult
}

func ConvCommitResult(msg *types.CommitResult) *InOutCommitResult {
	c := &InOutCommitResult{}
	c.Hash = base58.Encode(msg.Hash)
	c.Error = msg.Error
	c.Detail = msg.Detail
	return c
}

type InOutCommitResult struct {
	Hash   string
	Error  types.CommitStatus
	Detail string
}
