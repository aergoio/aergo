package enterprise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const SetConf = "setConf"
const SetAdmin = "setAdmin"
const EnableConf = "enableConf"

var ErrTxEnterpriseAdminIsNotSet = errors.New("admin is not set")

func ValidateEnterpriseTx(tx *types.TxBody, sender *state.V,
	scs *state.ContractState) (*EnterpriseContext, error) {
	var ci types.CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return nil, err
	}
	context := &EnterpriseContext{Call: &ci}
	switch ci.Name {
	case SetAdmin:
		if len(ci.Args) != 1 { //args[0] : key, args[1:] : values
			return nil, fmt.Errorf("invalid arguments in payload for SetAdmin: %s", ci.Args)
		}
		arg := ci.Args[0].(string)
		address := types.ToAddress(arg)
		if address == nil {
			return nil, fmt.Errorf("invalid arguments[0]: %s", ci.Args[0])
		}
		if err := checkAdmin(scs, sender.ID()); err != nil &&
			err != ErrTxEnterpriseAdminIsNotSet {
			return nil, err
		}

		context.Admin = address
	case SetConf:
		if len(ci.Args) <= 1 { //args[0] : key, args[1:] : values
			return nil, fmt.Errorf("invalid arguments in payload for setConf: %s", ci.Args)
		}
		for _, v := range ci.Args {
			arg, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("not string in payload for setConf : %s", ci.Args)
			}
			context.Args = append(context.Args, arg)
		}
		if err := checkAdmin(scs, sender.ID()); err != nil {
			return nil, err
		}
	case EnableConf:
		if len(ci.Args) != 2 { //args[0] : key, args[1] : true/false
			return nil, fmt.Errorf("invalid arguments in payload for enableConf: %s", ci.Args)
		}
		arg0, ok := ci.Args[0].(string)
		if !ok {
			return nil, fmt.Errorf("not string in payload for enableConf : %s", ci.Args)
		}
		context.Args = append(context.Args, arg0)
		_, ok = ci.Args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("not bool in payload for enableConf : %s", ci.Args)
		}
		if err := checkAdmin(scs, sender.ID()); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported call %s", ci.Name)
	}
	return context, nil
}

func checkAdmin(scs *state.ContractState, address []byte) error {
	admin, err := getAdmin(scs)
	if err != nil {
		return fmt.Errorf("could not get admin in enterprise contract")
	}
	if admin == nil {
		return ErrTxEnterpriseAdminIsNotSet
	}
	if !bytes.Equal(admin, address) {
		return fmt.Errorf("admin address not matched")
	}
	return nil
}
