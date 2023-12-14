package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvCommitResultList(msg *types.CommitResultList) *InOutCommitResultList {
	if msg == nil {
		return nil
	}
	c := &InOutCommitResultList{}
	c.Results = make([]*InOutCommitResult, len(msg.Results))
	for i, result := range msg.Results {
		c.Results[i] = ConvCommitResult(result)
	}
	return c
}

type InOutCommitResultList struct {
	Results []*InOutCommitResult `json:"results,omitempty"`
}

func ConvCommitResult(msg *types.CommitResult) *InOutCommitResult {
	return &InOutCommitResult{
		Hash:   base58.Encode(msg.Hash),
		Error:  msg.Error,
		Detail: msg.Detail,
	}
}

type InOutCommitResult struct {
	Hash   string             `json:"hash,omitempty"`
	Error  types.CommitStatus `json:"error,omitempty"`
	Detail string             `json:"detail,omitempty"`
}
