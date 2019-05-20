package enterprise

import (
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

func GetAdmin(r AccountStateReader) (*types.EnterpriseConfig, error) {
	scs, err := r.GetEnterpriseAccountState()
	if err != nil {
		return nil, err
	}
	admin, err := getAdmin(scs)
	if err != nil {
		return nil, err
	}
	ret := &types.EnterpriseConfig{Key: "admin", On: false}
	if admin != nil {
		ret.On = true
		ret.Values = append(ret.Values, types.EncodeAddress(admin))
	}
	return ret, nil
}
func setAdmin(scs *state.ContractState, address []byte) error {
	return scs.SetData([]byte("admin"), address)
}

func getAdmin(scs *state.ContractState) ([]byte, error) {
	return scs.GetData([]byte("admin"))
}
