package system

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

type TestAccountStateReader struct {
	Scs *state.ContractState
}

func (tas *TestAccountStateReader) GetSystemAccountState() (*state.ContractState, error) {
	if tas != nil && tas.Scs != nil {
		return tas.Scs, nil
	}
	return nil, fmt.Errorf("could not get system account state")
}

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

func TestProposalBPCount(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	//balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	//balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))

	sender2 := getSender(t, "AmNqJN2P1MA2Uc6X5byA4mDg2iuo95ANAyWCmd3LkZe4GhJkSyr4")
	sender3 := getSender(t, "AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7")
	sender.AddBalance(balance3)
	sender2.AddBalance(balance3)
	sender3.AddBalance(balance3)

	blockNo := uint64(0)
	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(scs, stakingTx.GetBody(), sender, receiver, blockNo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	stakingTx.Body.Account = sender2.ID()
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender2, receiver, blockNo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender2.Balance(), "sender.Balance() should be 2 after staking")

	stakingTx.Body.Account = sender3.ID()
	_, err = ExecuteSystemTx(scs, stakingTx.GetBody(), sender3, receiver, blockNo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender3.Balance(), "sender.Balance() should be 2 after staking")

	tx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender.ID(),
			Recipient: []byte(types.AergoSystem),
			Amount:    types.ProposalPrice.Bytes(),
			Payload:   []byte(`{"Name":"v1createProposal", "Args":["bpcount", "0","0","2","this vote is for the number of bp",[]]}`),
		},
	}
	events, err := ExecuteSystemTx(scs, tx.GetBody(), sender, receiver, blockNo)
	assert.NoError(t, err, "failed in creating proposal")
	assert.Equal(t, new(big.Int).Sub(balance2, types.ProposalPrice), sender.Balance(), "sender.Balance() should be 2 after creating proposal")
	assert.Equal(t, events[0].ContractAddress, types.AddressPadding([]byte(types.AergoSystem)), "check event")

	ar := &TestAccountStateReader{Scs: scs}
	InitDefaultBpCount(3)
	assert.Equal(t, lastBpCount, GetBpCount(ar), "check event")

	validCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteProposal", "Args":["bpcount", "13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(scs, validCandiTx.GetBody(), sender, receiver, blockNo)
	assert.NoError(t, err, "valid")

	assert.Equal(t, lastBpCount, GetBpCount(ar), "check bp")

	validCandiTx.Body.Account = sender2.ID()
	validCandiTx.Body.Payload = []byte(`{"Name":"v1voteProposal", "Args":["bpcount", "13", "17"]}`)

	_, err = ExecuteSystemTx(scs, validCandiTx.GetBody(), sender2, receiver, blockNo)
	assert.NoError(t, err, "valid")
	assert.Equal(t, 13, GetBpCount(ar), "check bp")
}
