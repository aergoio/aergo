package system

import (
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestAgendaSetGet(t *testing.T) {
	initTest(t)
	defer deinitTest()
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	originAgenda := &types.Agenda{
		Name:        "numofbp",
		Version:     "v0.1",
		Blockfrom:   1,
		Blockto:     2,
		Description: "the number of block producer",
		Candidates:  []string{"13", "23", "45"},
		Maxvote:     2,
	}
	_, err = getAgenda(scs, originAgenda.Name, originAgenda.Version)
	assert.NoError(t, err, "could not get agenda")
	err = setAgenda(scs, originAgenda)
	assert.NoError(t, err, "could not set agenda")
	testAgenda, err := getAgenda(scs, originAgenda.Name, originAgenda.Version)
	assert.NoError(t, err, "could not get agenda")
	assert.Equal(t, originAgenda.Name, testAgenda.Name, "agenda name")
	assert.Equal(t, originAgenda.Version, testAgenda.Version, "agenda version")
	assert.Equal(t, originAgenda.Description, testAgenda.Description, "agenda description")
	assert.Equal(t, originAgenda.Blockfrom, testAgenda.Blockfrom, "agenda blockfrom")
	assert.Equal(t, originAgenda.Blockto, testAgenda.Blockto, "agenda blockto")
	assert.Equal(t, originAgenda.Maxvote, testAgenda.Maxvote, "agenda max vote")

	originAgenda2 := &types.Agenda{
		Name:       "numofbp",
		Version:    "v0.1",
		Blockfrom:  1,
		Blockto:    2,
		Candidates: []string{"13", "23", "45"},
		Maxvote:    2,
	}
	err = setAgenda(scs, originAgenda2)
	assert.NoError(t, err, "could not get agenda")
	testAgenda2, err := getAgenda(scs, originAgenda2.Name, originAgenda2.Version)
	assert.NoError(t, err, "could not get agenda")
	assert.Equal(t, originAgenda2.Name, testAgenda2.Name, "agenda name")
	assert.Equal(t, originAgenda2.Version, testAgenda2.Version, "agenda version")
	assert.Equal(t, originAgenda2.Description, testAgenda2.Description, "agenda description")
	assert.Equal(t, originAgenda2.Blockfrom, testAgenda2.Blockfrom, "agenda max vote")
	assert.Equal(t, originAgenda2.Blockto, testAgenda2.Blockto, "agenda max vote")
	assert.Equal(t, originAgenda2.Maxvote, testAgenda2.Maxvote, "agenda max vote")
}

func buildAgendaPayload(t *testing.T, name, version string) (*types.CallInfo, []byte) {
	var ci types.CallInfo
	ci.Name = types.CreateAgenda
	agenda := &types.Agenda{
		Name:        name,
		Version:     version,
		Blockfrom:   1,
		Blockto:     2,
		Description: "the number of block producer",
		Candidates:  []string{"13", "23", "45"},
		Maxvote:     2,
	}
	//data, _ := json.Marshal(agenda)
	ci.Args = append(ci.Args, agenda)
	ret, _ := json.Marshal(ci)
	t.Log(string(ret))
	return &ci, ret
}
