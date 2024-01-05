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
	cr := &InOutCommitResult{
		Hash:   base58.Encode(msg.Hash),
		Error:  msg.Error,
		Detail: msg.Detail,
	}

	status, err := types.CommitStatus_name[int32(msg.Error)]
	if err && msg.Error != types.CommitStatus_TX_OK {
		cr.Message = status
	}

	return cr
}

type InOutCommitResult struct {
	Hash    string             `json:"hash,omitempty"`
	Error   types.CommitStatus `json:"error,omitempty"`
	Detail  string             `json:"detail,omitempty"`
	Message string             `json:"message,omitempty"`
}
