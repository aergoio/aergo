package system

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

const operatorKey = "operator"

type operatorCmd struct {
	*SystemContext
}

type Operators [][]byte

func (o Operators) IsExist(addr []byte) bool {
	for _, a := range o {
		if bytes.Equal(a, addr) {
			return true
		}
	}
	return false
}

func newOperatorCmd(ctx *SystemContext) (sysCmd, error) {
	return &operatorCmd{SystemContext: ctx}, nil
}

func (c *operatorCmd) run() (*types.Event, error) {
	var (
		scs       = c.scs
		receiver  = c.Receiver
		operators = c.Operators
	)
	if err := setOperators(scs, operators); err != nil {
		return nil, err
	}
	jsonArgs, err := json.Marshal(c.Args[0])
	if err != nil {
		return nil, err
	}
	return &types.Event{
		ContractAddress: receiver.ID(),
		EventIdx:        0,
		EventName:       c.op.ID(),
		JsonArgs:        string(jsonArgs),
	}, nil

}

func setOperators(scs *state.ContractState, addresses [][]byte) error {
	return scs.SetData([]byte(operatorKey), bytes.Join(addresses, []byte("")))
}

func getOperators(scs *state.ContractState) (Operators, error) {
	data, err := scs.GetData([]byte(operatorKey))
	if err != nil {
		return nil, err
	}
	var results [][]byte
	for i := 0; i < len(data); i += types.AddressLength {
		results = append(results, data[i:i+types.AddressLength])
	}
	return results, nil
}

func checkOperators(scs *state.ContractState, address []byte) (Operators, error) {
	operators, err := getOperators(scs)
	if err != nil {
		return nil, fmt.Errorf("could not get admin in enterprise contract")
	}
	if operators == nil {
		return nil, ErrTxSystemOperatorIsNotSet
	}
	if i := bytes.Index(bytes.Join(operators, []byte("")), address); i == -1 && i%types.AddressLength != 0 {
		return nil, fmt.Errorf("admin address not matched")
	}
	return operators, nil
}
