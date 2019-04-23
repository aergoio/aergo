package system

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var proposalListKey = []byte("proposallist")

type whereToVotes = [][]byte

func createProposal(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	proposal := context.Proposal
	amount := txBody.GetAmountBigInt()
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	if err := setProposal(scs, proposal); err != nil {
		return nil, err
	}
	log, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       context.Call.Name[2:],
		JsonArgs: `{"who":"` +
			types.EncodeAddress(txBody.Account) +
			`", "Proposal":` + string(log) + `}`,
	}, nil
}

//getProposal find proposal using id
func getProposal(scs *state.ContractState, id string) (*types.Proposal, error) {
	dataKey := types.GenProposalKey(id)
	data, err := scs.GetData([]byte(dataKey))
	if err != nil {
		return nil, fmt.Errorf("could not get proposal from contract state DB : %s", id)
	}
	return deserializeProposal(data), nil
}

func setProposal(scs *state.ContractState, proposal *types.Proposal) error {
	return scs.SetData(proposal.GetKey(), serializeProposal(proposal))
}

func serializeProposal(proposal *types.Proposal) []byte {
	data, err := json.Marshal(proposal)
	if err != nil {
		panic("could not marshal proposal")
	}
	return data
}

func deserializeProposal(data []byte) *types.Proposal {
	var proposal types.Proposal
	if err := json.Unmarshal(data, &proposal); err != nil {
		return nil
	}
	return &proposal
}

func getProposalHistory(scs *state.ContractState, address []byte) whereToVotes {
	key := append(proposalListKey, address...)
	return _getProposalHistory(scs, key)
}
func _getProposalHistory(scs *state.ContractState, key []byte) whereToVotes {
	data, err := scs.GetData(key)
	if err != nil {
		panic("could not get proposal history in contract state db")
	}
	if len(data) == 0 { //never vote before
		return nil
	}
	return deserializeProposalHistory(data)
}

func addProposalHistory(scs *state.ContractState, address []byte, proposal *types.Proposal) error {
	key := append(proposalListKey, address...)
	proposalHistory := _getProposalHistory(scs, key)
	proposalHistory = append(proposalHistory, proposal.GetKey())

	//unique
	filter := make(map[string]bool)
	var result whereToVotes
	for _, entryBytes := range proposalHistory {
		entry := string(entryBytes)
		if _, value := filter[entry]; !value {
			filter[entry] = true
			result = append(result, entryBytes)
		}
	}

	return scs.SetData(key, serializeProposalHistory(result))
}

func deserializeProposalHistory(data []byte) whereToVotes {
	return bytes.Split(data, []byte("/"))
}

func serializeProposalHistory(wtv whereToVotes) []byte {
	var data []byte
	for i, w := range wtv {
		if i != 0 {
			data = append(data, '/')
		}
		data = append(data, w...)
	}
	return data
}
