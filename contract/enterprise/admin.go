package enterprise

import (
	"bytes"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

const AdminsKey = "ADMINS"

func GetAdmin(r AccountStateReader) (*types.EnterpriseConfig, error) {
	scs, err := r.GetEnterpriseAccountState()
	if err != nil {
		return nil, err
	}
	admins, err := getAdmins(scs)
	if err != nil {
		return nil, err
	}
	ret := &types.EnterpriseConfig{Key: AdminsKey, On: false}
	if admins != nil {
		ret.On = true
		for _, admin := range admins {
			ret.Values = append(ret.Values, types.EncodeAddress(admin))
		}
	}
	return ret, nil
}
func setAdmins(scs *state.ContractState, addresses [][]byte) error {
	return scs.SetData([]byte(AdminsKey), bytes.Join(addresses, []byte("")))
}

func getAdmins(scs *state.ContractState) ([][]byte, error) {
	data, err := scs.GetData([]byte(AdminsKey))
	if err != nil {
		return nil, err
	}
	var results [][]byte
	for i := 0; i < len(data); i += types.AddressLength {
		results = append(results, data[i:i+types.AddressLength])
	}
	return results, nil
}
