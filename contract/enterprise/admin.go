package enterprise

import (
	"bytes"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/dbkey"
)

func GetAdmin(ecs *state.ContractState) (*types.EnterpriseConfig, error) {
	admins, err := getAdmins(ecs)
	if err != nil {
		return nil, err
	}
	ret := &types.EnterpriseConfig{Key: string(dbkey.EnterpriseAdmins()), On: false}
	if admins != nil {
		ret.On = true
		for _, admin := range admins {
			ret.Values = append(ret.Values, types.EncodeAddress(admin))
		}
	}
	return ret, nil
}
func setAdmins(scs *state.ContractState, addresses [][]byte) error {
	return scs.SetData(dbkey.EnterpriseAdmins(), bytes.Join(addresses, []byte("")))
}

func getAdmins(scs *state.ContractState) ([][]byte, error) {
	data, err := scs.GetData(dbkey.EnterpriseAdmins())
	if err != nil {
		return nil, err
	}
	var results [][]byte
	for i := 0; i < len(data); i += types.AddressLength {
		results = append(results, data[i:i+types.AddressLength])
	}
	return results, nil
}
