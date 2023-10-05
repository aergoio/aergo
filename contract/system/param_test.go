package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDefaultParams(t *testing.T) {
	// Staking minimum amount ( 10,000 aergo )
	stakingMin := DefaultParams.getParam(stakingMin.ID())
	assert.NotNilf(t, stakingMin, "stakingMin is not valid. check contract/system/param.go")
	assert.Equalf(t, "10000000000000000000000", stakingMin.String(), "StakingMinimum is not valid. check contract/system/param.go")

	// gas price ( 50 gaer )
	gasPrice := DefaultParams.getParam(gasPrice.ID())
	assert.NotNilf(t, gasPrice, "gasPrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "50000000000", gasPrice.String(), "GasPrice is not valid. check contract/system/param.go")

	// Proposal price ( 1 aergo )
	namePrice := DefaultParams.getParam(namePrice.ID())
	assert.NotNilf(t, namePrice, "namePrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "1000000000000000000", namePrice.String(), "ProposalPrice is not valid. check contract/system/param.go")
}
