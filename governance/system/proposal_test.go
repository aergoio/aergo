package system

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	testProposals *ProposalCatalog
)

func initProposalTest(t *testing.T) {
	testProposals = NewProposalCatalog(map[string]*Proposal{
		BpCount.ID(): {
			ID:             BpCount.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		StakingMin.ID(): {
			ID:             StakingMin.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		GasPrice.ID(): {
			ID:             GasPrice.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
		NamePrice.ID(): {
			ID:             NamePrice.ID(),
			Description:    "",
			Blockfrom:      0,
			Blockto:        0,
			MultipleChoice: 1,
			Candidates:     nil,
		},
	})
}

func deInitProposalTest() {
	testProposals = nil
}

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

	originProposal := &Proposal{
		ID:             "numofbp",
		Blockfrom:      1,
		Blockto:        2,
		Description:    "the number of block producer",
		Candidates:     []string{"13", "23", "45"},
		MultipleChoice: 2,
	}
	_, err := testProposals.GetProposal(originProposal.ID)
	assert.Error(t, err, "before set")
	testProposals.SetProposal(originProposal)
	testProposal, err := testProposals.GetProposal(originProposal.ID)
	assert.NoError(t, err, "could not get proposal")
	assert.Equal(t, originProposal.ID, testProposal.ID, "proposal name")
	assert.Equal(t, originProposal.Description, testProposal.Description, "proposal description")
	assert.Equal(t, originProposal.Blockfrom, testProposal.Blockfrom, "proposal blockfrom")
	assert.Equal(t, originProposal.Blockto, testProposal.Blockto, "proposal blockto")
	assert.Equal(t, originProposal.MultipleChoice, testProposal.MultipleChoice, "proposal max vote")

	originProposal2 := &Proposal{
		ID:             "numofbp",
		Blockfrom:      1,
		Blockto:        2,
		Candidates:     []string{"13", "23", "45"},
		MultipleChoice: 2,
	}
	testProposals.SetProposal(originProposal2)
	assert.NoError(t, err, "could not get proposal")
	testProposal2, err := testProposals.GetProposal(originProposal2.ID)
	assert.NoError(t, err, "could not get proposal")
	assert.Equal(t, originProposal2.ID, testProposal2.ID, "proposal name")
	assert.Equal(t, originProposal2.Description, testProposal2.Description, "proposal description")
	assert.Equal(t, originProposal2.Blockfrom, testProposal2.Blockfrom, "proposal max vote")
	assert.Equal(t, originProposal2.Blockto, testProposal2.Blockto, "proposal max vote")
	assert.Equal(t, originProposal2.MultipleChoice, testProposal2.MultipleChoice, "proposal max vote")
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

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	stakingTx.Body.Account = sender2.ID()
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender2.Balance(), "sender.Balance() should be 2 after staking")

	stakingTx.Body.Account = sender3.ID()
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender3.Balance(), "sender.Balance() should be 2 after staking")

	validCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "before v2")

	blockInfo.No++ //set v2
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "valid")

	assert.Equal(t, 3, TestParams.GetBpCount(), "check bp")

	validCandiTx.Body.Account = sender2.ID()
	validCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13"]}`)

	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "valid")
	assert.Equal(t, 13, TestParams.GetBpCount(), "check bp")
}

func TestFailProposals(t *testing.T) {
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

	blockInfo := &types.BlockHeaderInfo{No: uint64(0)}
	stakingTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Amount:  balance1.Bytes(),
			Payload: buildStakingPayload(true),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err := ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender.Balance(), "sender.Balance() should be 1 after staking")

	stakingTx.Body.Account = sender2.ID()
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender2.Balance(), "sender.Balance() should be 2 after staking")

	stakingTx.Body.Account = sender3.ID()
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, stakingTx.GetBody(), sender3, receiver, blockInfo)
	assert.NoError(t, err, "could not execute system tx")
	assert.Equal(t, balance2, sender3.Balance(), "sender.Balance() should be 2 after staking")

	validCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "before v2")

	blockInfo.No++ //set v2
	blockInfo.ForkVersion = config.AllEnabledHardforkConfig.Version(blockInfo.No)

	invalidCandiTx := &types.Tx{
		Body: &types.TxBody{
			Account: sender.ID(),
			Payload: []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "0"]}`),
			Type:    types.TxType_GOVERNANCE,
		},
	}

	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, invalidCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "invalid range")

	invalidCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "101"]}`)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, invalidCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "invalid range")

	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "valid")

	assert.Equal(t, 3, TestParams.GetBpCount(), "check bp")

	validCandiTx.Body.Account = sender2.ID()
	validCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["bpcount", "13"]}`)

	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "valid")
	assert.Equal(t, 13, TestParams.GetBpCount(), "check bp")

	invalidCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["gasprice", "500000000000000000000000001"]}`)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, invalidCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.Error(t, err, "invalid range")

	invalidCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["gasprice", "5000aergo"]}`)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, invalidCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.EqualError(t, err, "include invalid number", "invalid number")

	validCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["gasprice", "101"]}`)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender, receiver, blockInfo)
	assert.NoError(t, err, "valid")
	assert.Equal(t, TestParams.params[GasPrice.ID()], TestParams.GetGasPrice(), "check gas price")

	validCandiTx.Body.Payload = []byte(`{"Name":"v1voteDAO", "Args":["gasprice", "101"]}`)
	_, err = ExecuteSystemTx(testProposals, testVotingCatalog, scs, validCandiTx.GetBody(), sender2, receiver, blockInfo)
	assert.NoError(t, err, "valid")
	assert.Equal(t, big.NewInt(101), TestParams.GetGasPrice(), "check gas price")
}
