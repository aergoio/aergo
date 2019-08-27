package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestOperatorFail(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	//balance0_5 := new(big.Int).Div(types.StakingMinimum, big.NewInt(2))
	balance1 := types.StakingMinimum
	//balance1_5 := new(big.Int).Add(balance1, balance0_5)
	balance2 := new(big.Int).Mul(balance1, big.NewInt(2))
	balance3 := new(big.Int).Mul(balance1, big.NewInt(3))

	sender.AddBalance(balance3)
	sender2 := getSender(t, "AmNqJN2P1MA2Uc6X5byA4mDg2iuo95ANAyWCmd3LkZe4GhJkSyr4")
	sender2.AddBalance(balance3)

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

	operatorTx := &types.Tx{
		Body: &types.TxBody{
			Account:   sender2.ID(),
			Recipient: []byte(types.AergoSystem),
			Amount:    types.ProposalPrice.Bytes(),
			Payload:   []byte(`{"Name":"v1addOperator", "Args":["AmNqJN2P1MA2Uc6X5byA4mDg2iuo95ANAyWCmd3LkZe4GhJkSyr4"]}`),
		},
	}
	_, err = ExecuteSystemTx(scs, operatorTx.GetBody(), sender2, receiver, blockNo)
	assert.NoError(t, err, "could not set system operator")
}
