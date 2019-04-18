package system

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var agendaListKey = []byte("agendalist")

type whereToVotes = [][]byte

//voteAgenda is vote to specific agenda which is identified with agenda name and version. before call this function, should validate transaction first.
func voteAgenda(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	return voting(txBody, sender, receiver, scs, blockNo, context)
}

func createAgenda(txBody *types.TxBody, sender, receiver *state.V, scs *state.ContractState,
	blockNo types.BlockNo, context *SystemContext) (*types.Event, error) {
	agenda := context.Agenda
	amount := txBody.GetAmountBigInt()
	sender.SubBalance(amount)
	receiver.AddBalance(amount)
	addAgendaHistory(scs, sender.ID(), agenda)
	return nil, setAgenda(scs, agenda)
}

//getAgenda find agenda using name and version
func getAgenda(scs *state.ContractState, name, version string) (*types.Agenda, error) {
	dataKey := types.GenAgendaKey(name, version)
	data, err := scs.GetData([]byte(dataKey))
	if err != nil {
		return nil, fmt.Errorf("could not get agenda from contract state DB : %s %s", name, version)
	}
	return deserializeAgenda(data), nil
}

func setAgenda(scs *state.ContractState, agenda *types.Agenda) error {
	return scs.SetData(agenda.GetKey(), serializeAgenda(agenda))
}

func serializeAgenda(agenda *types.Agenda) []byte {
	data, err := json.Marshal(agenda)
	if err != nil {
		panic("could not marshal agenda")
	}
	return data
}

func deserializeAgenda(data []byte) *types.Agenda {
	var agenda types.Agenda
	if err := json.Unmarshal(data, &agenda); err != nil {
		return nil
	}
	return &agenda
}

func getAgendaHistory(scs *state.ContractState, address []byte) whereToVotes {
	key := append(agendaListKey, address...)
	return _getAgendaHistory(scs, key)
}
func _getAgendaHistory(scs *state.ContractState, key []byte) whereToVotes {
	data, err := scs.GetData(key)
	if err != nil {
		panic("could not get agenda list")
	}
	return deserializeAgendaHistory(data)
}

func addAgendaHistory(scs *state.ContractState, address []byte, agenda *types.Agenda) error {
	key := append(agendaListKey, address...)
	agendaHistory := _getAgendaHistory(scs, key)
	agendaHistory = append(agendaHistory, agenda.GetKey())
	return scs.SetData(key, serializeAgendaHistory(agendaHistory))
}

func deserializeAgendaHistory(data []byte) whereToVotes {
	return bytes.Split(data, []byte("|"))
}

func serializeAgendaHistory(wtv whereToVotes) []byte {
	var data []byte
	for i, w := range wtv {
		if i != 0 {
			data = append(data, '|')
		}
		data = append(data, w...)
	}
	return data
}
