package system

import (
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/types"
)

var (
	TestParams *Parameters
)

func initParamsTest(t *testing.T) {
	TestParams = NewParameters()
	TestParams.SetBpCount(big.NewInt(3))
	TestParams.SetStakingMinimum(types.StakingMinimum)
	TestParams.SetGasPrice(types.NewAmount(50, types.Gaer))
	TestParams.SetNamePrice(types.NewAmount(1, types.Aergo))
}

/*
func TestValidateDefaultParams(t *testing.T) {
	// Staking minimum amount ( 10,000 aergo )
	stakingMin, ok := DefaultParams[StakingMin.ID()]
	assert.Truef(t, ok, "stakingMin is not valid. check contract/system/param.go")
	assert.Equalf(t, "10000000000000000000000", stakingMin.String(), "StakingMinimum is not valid. check contract/system/param.go")

	// gas price ( 50 gaer )
	gasPrice, ok := DefaultParams[GasPrice.ID()]
	assert.Truef(t, ok, "gasPrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "50000000000", gasPrice.String(), "GasPrice is not valid. check contract/system/param.go")

	// Proposal price ( 1 aergo )
	namePrice := DefaultParams[NamePrice.ID()]
	assert.Truef(t, ok, "namePrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "1000000000000000000", namePrice.String(), "ProposalPrice is not valid. check contract/system/param.go")
}
*/
