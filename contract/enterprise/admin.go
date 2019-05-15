package enterprise

import (
	"github.com/aergoio/aergo/state"
)

func setAdmin(scs *state.ContractState, address []byte) error {
	return scs.SetData([]byte("admin"), address)
}

func getAdmin(scs *state.ContractState) ([]byte, error) {
	return scs.GetData([]byte("admin"))
}
