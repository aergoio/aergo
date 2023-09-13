package governance

import (
	"math/big"

	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/types"
)

type SystemSnapshot struct {
	systemParams    system.Parameters
	nameParams      map[string]interface{}
	votingPowerRank *system.Vpr
}

func (ss *SystemSnapshot) Init(getter system.DataGetter) error {
	var err error
	ss.systemParams = map[string]*big.Int{
		system.StakingMin.ID(): types.StakingMinimum,
		system.GasPrice.ID():   types.NewAmount(50, types.Gaer), // 50 gaer
		system.NamePrice.ID():  types.NewAmount(1, types.Aergo), // 1 aergo
	}

	ss.votingPowerRank, err = system.LoadVpr(getter)
	if err != nil {
		return err
	}
	return nil
}

func (ss *SystemSnapshot) Copy() *SystemSnapshot {
	new := &SystemSnapshot{
		systemParams: map[string]*big.Int{},
		nameParams:   map[string]interface{}{},
	}
	// TODO
	// for k, v := range ss.systemParams {
	// new.systemParams[k] = big.NewInt(0).Set(v)
	// }
	return new
}
