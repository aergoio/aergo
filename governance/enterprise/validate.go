package enterprise

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
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
	scs *state.ContractState, blockNo types.BlockNo) (*EnterpriseContext, error) {
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
		} else if ci.Name == RemoveAdmin {
			if !context.IsAdminExist(address) {
				return nil, fmt.Errorf("admins is not exist : %s", ci.Args[0])
			}
			conf, err := getConf(scs, []byte(AccountWhite))
			if err != nil {
				return nil, err
			}
			if conf != nil && conf.On {
				for _, v := range conf.Values {
					if arg == v {
						return nil, fmt.Errorf("admin is in the account whitelist: %s", ci.Args[0])
					}
				}
			}
		}

	case SetConf:
		if len(ci.Args) <= 1 { //args[0] : key, args[1:] : values
			return nil, fmt.Errorf("invalid arguments in payload for setConf: %s", ci.Args)
		}
		if err := checkArgs(context, &ci); err != nil {
			return nil, err
		}
		key := genKey([]byte(context.Args[0]))
		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		context.Admins = admins
		if context.Conf, err = setConfValues(scs, key, context.Args[1:]); err != nil {
			return nil, err
		}

		if conf, err := getConf(scs, []byte(context.Args[0])); err == nil && conf != nil {
			if err := conf.Validate(key, context); err != nil {
				return nil, err
			}
		}

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
		conf, err := getConf(scs, []byte(context.Args[0]))
		if err != nil {
			return nil, err
		}
		if conf == nil {
			conf = &Conf{On: false}
		}
		context.Conf = conf
		context.Admins = admins

		key := genKey([]byte(context.Args[0]))
		if ci.Name == AppendConf {
			if context.HasConfValue(context.Args[1]) {
				return nil, fmt.Errorf("already included config value : %v", context.Args)
			}
			conf.AppendValue(context.Args[1])
		} else if ci.Name == RemoveConf {
			if !context.HasConfValue(context.Args[1]) {
				return nil, fmt.Errorf("value not exist : %v", context.Args)
			}
			conf.RemoveValue(context.Args[1])
		}

		if err := conf.Validate(key, context); err != nil {
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
		if _, ok := enterpriseKeyDict[strings.ToUpper(ci.Args[0].(string))]; !ok {
			return nil, fmt.Errorf("not allowed key : %s", ci.Args[0])
		}
		context.Args = append(context.Args, arg0)
		key := genKey([]byte(arg0))
		value, ok := ci.Args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("not bool in payload for enableConf : %s", ci.Args)
		}
		admins, err := checkAdmin(scs, sender.ID())
		if err != nil {
			return nil, err
		}
		conf, err := enableConf(scs, key, value)
		if err != nil {
			return nil, err
		}

		context.Admins = admins
		context.Conf = conf

		if err := conf.Validate(key, context); err != nil {
			return nil, err
		}

	case ChangeCluster:
		if !consensus.UseRaft() {
			return nil, ErrNotSupportedMethod
		}

		cc, err := ValidateChangeCluster(ci, blockNo)
		if err != nil {
			return nil, err
		}

		context.ArgsAny = append(context.ArgsAny, cc)

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

type checkArgsFunc func(string) error

func checkArgs(context *EnterpriseContext, ci *types.CallInfo) error {
	key := strings.ToUpper(ci.Args[0].(string))
	if _, ok := enterpriseKeyDict[key]; !ok {
		return fmt.Errorf("not allowed key : %s", ci.Args[0])
	}

	unique := map[string]int{}
	var op checkArgsFunc
	switch key {
	case P2PWhite, P2PBlack:
		op = checkP2PBlackWhite
	case AccountWhite:
		op = checkAccountWhite
	case RPCPermissions:
		op = checkRPCPermissions
	default:
		op = checkNone
	}

	for i, v := range ci.Args {
		arg, ok := v.(string)
		if !ok {
			return fmt.Errorf("not string in payload for setting conf : %s", ci.Args)
		}
		if strings.Contains(arg, "\\") {
			return fmt.Errorf("not allowed charactor in %s", arg)
		}
		if unique[arg] != 0 {
			return fmt.Errorf("the request has duplicate arguments %s", arg)
		}
		unique[arg]++
		context.Args = append(context.Args, arg)
		if i == 0 { //it's key
			continue
		}
		if err := op(arg); err != nil {
			return err
		}
	}

	return nil
}

func checkP2PBlackWhite(v string) error {
	// v must be json object. e.g. {"peerid":"16Uiu2HAmPZE7gT1hF2bjpg1UVH65xyNUbBVRf3mBFBJpz3tgLGGt", "address":"", "cidr":"172.21.3.35/24" } , which address and cidr cannot be set in same time.
	if _, err := types.ParseListEntry(v); err != nil {
		return fmt.Errorf("invalid p2p whitelist %s", v)
	}
	return nil
}

func checkAccountWhite(v string) error {
	if _, err := types.DecodeAddress(v); err != nil {
		return fmt.Errorf("invalid account %s", v)
	}
	return nil
}

func checkRPCPermissions(v string) error {
	values := strings.Split(v, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid RPC permission %s", v)
	}

	if _, err := base64.StdEncoding.DecodeString(values[0]); err != nil {
		return fmt.Errorf("invalid RPC cert %s", v)
	}

	return nil
}

func checkNone(v string) error {
	return nil
}
