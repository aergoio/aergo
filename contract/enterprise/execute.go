package enterprise

import (
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

type EnterpriseContext struct {
	Call  *types.CallInfo
	Args  []string
	Admin []byte
}

func ExecuteEnterpriseTx(scs *state.ContractState, txBody *types.TxBody,
	sender *state.V) ([]*types.Event, error) {
	context, err := ValidateEnterpriseTx(txBody, sender, scs)
	if err != nil {
		return nil, err
	}
	switch context.Call.Name {
	case SetAdmin:
		err := setAdmin(scs, context.Admin)
		if err != nil {
			return nil, err
		}
	case SetConf:
		key := []byte(context.Args[0])
		err := setConf(scs, key, &Conf{
			On:     false,
			Values: context.Args[1:],
		})
		if err != nil {
			return nil, err
		}
	case EnableConf:
		key := []byte(context.Args[0])
		if err := enableConf(scs, key, context.Call.Args[1].(bool)); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported call in enterprise contract")
	}
	return nil, nil
}
