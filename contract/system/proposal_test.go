package system

import (
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestProposalSetGet(t *testing.T) {
	initTest(t)
	defer deinitTest()
	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")
	originProposal := &types.Proposal{
		Id:          "numofbp",
		Blockfrom:   1,
		Blockto:     2,
		Description: "the number of block producer",
		Candidates:  []string{"13", "23", "45"},
		Maxvote:     2,
	}
	_, err = getProposal(scs, originProposal.Id)
	assert.NoError(t, err, "could not get proposal")
	err = setProposal(scs, originProposal)
	assert.NoError(t, err, "could not set proposal")
	testProposal, err := getProposal(scs, originProposal.Id)
	assert.NoError(t, err, "could not get proposal")
	assert.Equal(t, originProposal.Id, testProposal.Id, "proposal name")
	assert.Equal(t, originProposal.Description, testProposal.Description, "proposal description")
	assert.Equal(t, originProposal.Blockfrom, testProposal.Blockfrom, "proposal blockfrom")
	assert.Equal(t, originProposal.Blockto, testProposal.Blockto, "proposal blockto")
	assert.Equal(t, originProposal.Maxvote, testProposal.Maxvote, "proposal max vote")

	originProposal2 := &types.Proposal{
		Id:         "numofbp",
		Blockfrom:  1,
		Blockto:    2,
		Candidates: []string{"13", "23", "45"},
		Maxvote:    2,
	}
	err = setProposal(scs, originProposal2)
	assert.NoError(t, err, "could not get proposal")
	testProposal2, err := getProposal(scs, originProposal2.Id)
	assert.NoError(t, err, "could not get proposal")
	assert.Equal(t, originProposal2.Id, testProposal2.Id, "proposal name")
	assert.Equal(t, originProposal2.Description, testProposal2.Description, "proposal description")
	assert.Equal(t, originProposal2.Blockfrom, testProposal2.Blockfrom, "proposal max vote")
	assert.Equal(t, originProposal2.Blockto, testProposal2.Blockto, "proposal max vote")
	assert.Equal(t, originProposal2.Maxvote, testProposal2.Maxvote, "proposal max vote")
}

func buildProposalPayload(t *testing.T, name, version string) (*types.CallInfo, []byte) {
	var ci types.CallInfo
	ci.Name = types.CreateProposal
	proposal := &types.Proposal{
		Id:          name,
		Blockfrom:   1,
		Blockto:     2,
		Description: "the number of block producer",
		Candidates:  []string{"13", "23", "45"},
		Maxvote:     2,
	}
	//data, _ := json.Marshal(proposal)
	ci.Args = append(ci.Args, proposal)
	ret, _ := json.Marshal(ci)
	t.Log(string(ret))
	return &ci, ret
}
