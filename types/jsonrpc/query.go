package jsonrpc

import (
	"github.com/aergoio/aergo/v2/types"
)

func ConvQueryContract(msg *types.SingleBytes) *InOutQueryContract {
	if msg == nil {
		return nil
	}
	q := &InOutQueryContract{}
	q.Result = string(msg.Value)
	return q
}

type InOutQueryContract struct {
	Result string `json:"result"`
}

func ConvContractVarProof(msg *types.ContractVarProof) *InOutContractVarProof {
	if msg == nil {
		return nil
	}
	ap := &InOutContractVarProof{}
	ap.Value = string(msg.GetValue())
	ap.Included = msg.GetInclusion()
	ap.MerkleProofLength = len(msg.GetAuditPath())
	ap.Height = msg.GetHeight()

	return ap
}

type InOutContractVarProof struct {
	Value             string `json:"value,omitempty"`
	Included          bool   `json:"included,omitempty"`
	MerkleProofLength int    `json:"merkleprooflength,omitempty"`
	Height            uint32 `json:"height,omitempty"`
}

func ConvQueryContractState(msg *types.StateQueryProof) *InOutQueryContractState {
	if msg == nil {
		return nil
	}
	qcs := &InOutQueryContractState{}
	qcs.ContractProof = ConvStateAndPoof(msg.ContractProof)

	qcs.VarProofs = make([]*InOutContractVarProof, len(msg.VarProofs))
	for i, varProof := range msg.VarProofs {
		qcs.VarProofs[i] = ConvContractVarProof(varProof)
	}

	return qcs
}

type InOutQueryContractState struct {
	ContractProof *InOutStateAndPoof       `json:"contractProof,omitempty"`
	VarProofs     []*InOutContractVarProof `json:"varProofs,omitempty"`
}
