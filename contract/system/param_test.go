package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDefaultParams(t *testing.T) {
	// Staking minimum amount ( 10,000 aergo )
	stakingMin, ok := DefaultParams[stakingMin.ID()]
	assert.Truef(t, ok, "stakingMin is not valid. check contract/system/param.go")
	assert.Equalf(t, "10000000000000000000000", stakingMin.String(), "StakingMinimum is not valid. check contract/system/param.go")

	// gas price ( 50 gaer )
	gasPrice, ok := DefaultParams[gasPrice.ID()]
	assert.Truef(t, ok, "gasPrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "50000000000", gasPrice.String(), "GasPrice is not valid. check contract/system/param.go")

	// Proposal price ( 1 aergo )
	namePrice := DefaultParams[namePrice.ID()]
	assert.Truef(t, ok, "namePrice is not valid. check contract/system/param.go")
	assert.Equalf(t, "1000000000000000000", namePrice.String(), "ProposalPrice is not valid. check contract/system/param.go")
}
