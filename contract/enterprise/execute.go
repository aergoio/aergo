package enterprise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/consensus"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var (
	entLogger *log.Logger

	ErrNotSupportedMethod = errors.New("Not supported Enterprise Tx")
)

type EnterpriseContext struct {
	Call    *types.CallInfo
	Args    []string
	ArgsAny []interface{}
	Admins  [][]byte
	Conf    *Conf
}

func init() {
	entLogger = log.NewLogger("enterprise")
}

func (e *EnterpriseContext) IsAdminExist(addr []byte) bool {
	for _, a := range e.Admins {
		if bytes.Equal(a, addr) {
			return true
		}
	}
	return false
}

func (e *EnterpriseContext) HasConfValue(value string) bool {
	if e.Conf != nil {
		for _, v := range e.Conf.Values {
			if v == value {
				return true
			}
		}
	}
	return false
}

func ExecuteEnterpriseTx(ccc consensus.ChainConsensusCluster, scs *state.ContractState, txBody *types.TxBody,
	sender *state.V) ([]*types.Event, error) {
	context, err := ValidateEnterpriseTx(txBody, sender, scs)
	if err != nil {
		return nil, err
	}
	var events []*types.Event
	switch context.Call.Name {
	case AppendAdmin:
		requestAddress := types.ToAddress(context.Args[0])
		err := setAdmins(scs,
			append(context.Admins, requestAddress))
		if err != nil {
			return nil, err
		}
	case RemoveAdmin:
		for i, v := range context.Admins {
			if bytes.Equal(v, types.ToAddress(context.Call.Args[0].(string))) {
				context.Admins = append(context.Admins[:i], context.Admins[i+1:]...)
				break
			}
		}
		err := setAdmins(scs, context.Admins)
		if err != nil {
			return nil, err
		}
	case SetConf, AppendConf, RemoveConf:
		key := context.Args[0]
		err = setConf(scs, []byte(key), context.Conf)
		if err != nil {
			return nil, err
		}
		events, err = createSetEvent(txBody.Recipient, key, context.Conf.Values)
		if err != nil {
			return nil, err
		}
	case EnableConf:
		key := context.Args[0]
		err = setConf(scs, []byte(key), context.Conf)
		if err != nil {
			return nil, err
		}
		jsonArgs, err := json.Marshal(context.Call.Args[1])
		if err != nil {
			return nil, err
		}
		events = append(events, &types.Event{
			ContractAddress: txBody.Recipient,
			EventName:       "Enable " + string(key),
			EventIdx:        0,
			JsonArgs:        string(jsonArgs),
		})
	case ChangeCluster:
		if ccc == nil {
			return nil, ErrNotSupportedMethod
		}
		ccReq, ok := context.ArgsAny[0].(*types.MembershipChange)
		if !ok {
			return nil, fmt.Errorf("invalid argument of cluster change request")
		}
		//entLogger.Info().Str("req", ccReq.ToString()).Msg("Enterprise tx: cluster change request")

		if err := ccc.RequestConfChange(ccReq); err != nil && err != consensus.ErrorMembershipChangeSkip {
			return nil, err
		}

		jsonArgs, err := json.Marshal(context.Call.Args[0])
		if err != nil {
			return nil, err
		}

		entLogger.Debug().Str("jsonarg", string(jsonArgs)).Msg("make event")
		/*
			events = append(events, &types.Event{
				ContractAddress: txBody.Recipient,
				EventName:       "ChangeCluster ",
				EventIdx:        0,
				JsonArgs:        string(jsonArgs),
			})*/
	default:
		return nil, fmt.Errorf("unsupported call in enterprise contract")
	}
	return events, nil
}

func createSetEvent(addr []byte, name string, v []string) ([]*types.Event, error) {
	jsonArgs, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return []*types.Event{
		&types.Event{
			ContractAddress: addr,
			EventName:       "Set " + name,
			EventIdx:        0,
			JsonArgs:        string(jsonArgs),
		},
	}, nil
}
