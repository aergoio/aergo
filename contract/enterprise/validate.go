package enterprise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const SetConf = "setConf"
const AppendConf = "appendConf"
const RemoveConf = "removeConf"
const AppendAdmin = "appendAdmin"
const RemoveAdmin = "removeAdmin"
const EnableConf = "enableConf"
const DisableConf = "disableConf"
const ChangeCluster = "changeCluster"

var ErrTxEnterpriseAdminIsNotSet = errors.New("admin is not set")

func ValidateEnterpriseTx(tx *types.TxBody, sender *state.V,
	scs *state.ContractState) (*EnterpriseContext, error) {
	var ci types.CallInfo
	if err := json.Unmarshal(tx.Payload, &ci); err != nil {
		return nil, err
	}
	context := &EnterpriseContext{Call: &ci}
	switch ci.Name {
	case AppendAdmin, RemoveAdmin:
		if len(ci.Args) != 1 { //args[0] : encoded admin address
			return nil, fmt.Errorf("invalid arguments in payload for SetAdmin: %s", ci.Args)
		}
		arg := ci.Args[0].(string)
		context.Args = append(context.Args, arg)

		address := types.ToAddress(arg)
		if len(address) == 0 {
			return nil, fmt.Errorf("invalid arguments[0]: %s", ci.Args[0])
		}
		admins, err := checkAdmin(scs, sender.ID())
		if err != nil &&
			err != ErrTxEnterpriseAdminIsNotSet {
			return nil, err
		}
		context.Admins = admins
		if ci.Name == AppendAdmin && context.IsAdminExist(address) {
			return nil, fmt.Errorf("already exist admin: %s", ci.Args[0])
		} else if ci.Name == RemoveAdmin && !context.IsAdminExist(address) {
			return nil, fmt.Errorf("admins is not exist : %s", ci.Args[0])
		}

	case SetConf:
		if len(ci.Args) <= 1 { //args[0] : key, args[1:] : values
			return nil, fmt.Errorf("invalid arguments in payload for setConf: %s", ci.Args)
		}
		if err := checkArgs(context, &ci); err != nil {
			return nil, err
		}

		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		context.Admins = admins

	case AppendConf, RemoveConf:
		if len(ci.Args) != 2 { //args[0] : key, args[1] : a value
			return nil, fmt.Errorf("invalid arguments in payload for %s : %s", ci.Name, ci.Args)
		}
		if err := checkArgs(context, &ci); err != nil {
			return nil, err
		}
		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		old, err := getConf(scs, []byte(context.Args[0]))
		if err != nil {
			return nil, err
		}
		context.Old = old
		if ci.Name == AppendConf && context.IsOldConfValue(context.Args[1]) {
			return nil, fmt.Errorf("already included config value : %v", context.Args)
		} else if ci.Name == RemoveConf && !context.IsOldConfValue(context.Args[1]) {
			return nil, fmt.Errorf("value not exist : %v", context.Args)
		}
		context.Admins = admins

	case EnableConf:
		if len(ci.Args) != 2 { //args[0] : key, args[1] : true/false
			return nil, fmt.Errorf("invalid arguments in payload for enableConf: %s", ci.Args)
		}
		arg0, ok := ci.Args[0].(string)
		if !ok {
			return nil, fmt.Errorf("not string in payload for enableConf : %s", ci.Args)
		}
		if strings.ToLower(arg0) == "admin" {
			return nil, fmt.Errorf("not allowed key : %s", ci.Args[0])
		}
		context.Args = append(context.Args, arg0)
		_, ok = ci.Args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("not bool in payload for enableConf : %s", ci.Args)
		}
		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		context.Admins = admins
	case ChangeCluster:
		var (
			arg0, arg1 string
			ok         bool
			changeReq  *types.MembershipChange
			err        error
		)

		if len(ci.Args) != 2 { //args[0] : key, args[1] : true/false
			return nil, fmt.Errorf("invalid arguments in payload for ChangeCluster: %s", ci.Args)
		}
		arg0, ok = ci.Args[0].(string)
		if !ok {
			return nil, fmt.Errorf("not string in command for ChangeCluster : %v", ci.Args[0])
		}

		arg1, ok = ci.Args[1].(string)
		if !ok {
			return nil, fmt.Errorf("not string in member meta for ChangeCluster : %v", ci.Args[1])
		}

		if changeReq, err = parseChangeCluster(arg0, arg1); err != nil {
			return nil, fmt.Errorf("failed to parse ChangeCluster : %v", ci.Args)
		}

		context.ArgsAny = append(context.ArgsAny, changeReq)

		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		context.Admins = admins
	default:
		return nil, fmt.Errorf("unsupported call %s", ci.Name)
	}
	return context, nil
}

func parseChangeCluster(cmdStr string, meta string) (*types.MembershipChange, error) {
	const (
		CmdMembershipAdd    = "add"
		CmdMembershipRemove = "remove"
	)

	cmdStr = strings.ToLower(cmdStr)
	var cmd types.MembershipChangeType

	switch cmdStr {
	case CmdMembershipAdd:
		cmd = types.MembershipChangeType_ADD_MEMBER
	case CmdMembershipRemove:
		cmd = types.MembershipChangeType_REMOVE_MEMBER
	default:
		return nil, fmt.Errorf("not allowed command : %s", cmdStr)
	}

	var attr = types.JsonMemberAttr{}

	if err := json.Unmarshal([]byte(meta), &attr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta of ChangeCluster : %s, %v", meta, err)
	}

	switch cmd {
	case types.MembershipChangeType_ADD_MEMBER:
		attr.ID = 0
	case types.MembershipChangeType_REMOVE_MEMBER:
		if attr.ID == 0 {
			return nil, fmt.Errorf("id must exist to remove member from cluster: %s", meta)
		}
	}

	return &types.MembershipChange{Type: cmd, Attr: attr.ToMemberAttr()}, nil
}

func checkAdmin(scs *state.ContractState, address []byte) ([][]byte, error) {
	admins, err := getAdmins(scs)
	if err != nil {
		return nil, fmt.Errorf("could not get admin in enterprise contract")
	}
	if admins == nil {
		return nil, ErrTxEnterpriseAdminIsNotSet
	}
	if i := bytes.Index(bytes.Join(admins, []byte("")), address); i == -1 && i%types.AddressLength != 0 {
		return nil, fmt.Errorf("admin address not matched")
	}
	return admins, nil
}

func checkArgs(context *EnterpriseContext, ci *types.CallInfo) error {
	if strings.ToUpper(ci.Args[0].(string)) == adminsKey {
		return fmt.Errorf("not allowed key : %s", ci.Args[0])
	}

	unique := map[string]int{}
	for _, v := range ci.Args {
		arg, ok := v.(string)
		if !ok {
			return fmt.Errorf("not string in payload for setConf : %s", ci.Args)
		}
		if strings.Contains(arg, "\\") {
			return fmt.Errorf("not allowed charactor in %s", arg)
		}
		if unique[arg] != 0 {
			return fmt.Errorf("the request has duplicate arguments %s", arg)
		}
		unique[arg]++
		context.Args = append(context.Args, arg)
	}
	return nil
}
