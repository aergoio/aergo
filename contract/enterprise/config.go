package enterprise

import (
	"fmt"
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var confPrefix = []byte("conf\\")

const (
	RPCPermissions = "RPCPERMISSIONS"
	P2PWhite       = "P2PWHITE"
)

//enterpriseKeyDict is represent allowed key list and used when validate tx, int values are meaningless.
var enterpriseKeyDict = map[string]int{
	RPCPermissions: 1,
	P2PWhite:       2,
}

type Conf struct {
	On     bool
	Values []string
}

func (c *Conf) RemoveValue(r string) {
	for i, v := range c.Values {
		if v == r {
			c.Values = append(c.Values[:i], c.Values[i+1:]...)
			break
		}
	}
}

func (c *Conf) Validate(key []byte) error {
	if !c.On {
		return nil
	}
	strKey := string(key)
	switch strKey {
	case RPCPermissions:
		for _, v := range c.Values {
			if strings.Contains(strings.ToUpper(strings.Split(v, ":")[1]), "W") {
				return nil
			}
		}
		return fmt.Errorf("the values of %s should have at least one write permission", strKey)
	case P2PWhite:
	default:
		return fmt.Errorf("could not validate key(%s)", strKey)
	}
	return nil
}

// AccountStateReader is an interface for getting a enterprise account state.
type AccountStateReader interface {
	GetEnterpriseAccountState() (*state.ContractState, error)
}

func GetConf(r AccountStateReader, key string) (*types.EnterpriseConfig, error) {
	scs, err := r.GetEnterpriseAccountState()
	if err != nil {
		return nil, err
	}
	conf, err := getConf(scs, []byte(key))
	if err != nil {
		return nil, err
	}
	ret := &types.EnterpriseConfig{Key: key}
	if conf != nil {
		ret.On = conf.On
		ret.Values = conf.Values
	}
	return ret, nil
}

func enableConf(scs *state.ContractState, key []byte, value bool) error {
	conf, err := getConf(scs, key)
	if err != nil {
		return err
	}
	if conf != nil {
		conf.On = value
	} else {
		conf = &Conf{On: value}
	}

	return setConf(scs, key, conf)
}

func getConf(scs *state.ContractState, key []byte) (*Conf, error) {
	data, err := scs.GetData(append(confPrefix, genKey(key)...))
	if err != nil || data == nil {
		return nil, err
	}
	return deserializeConf(data), err
}

func setConfValues(scs *state.ContractState, key []byte, in *Conf) error {
	conf, err := getConf(scs, key)
	if err != nil {
		return err
	}
	if conf != nil {
		conf.Values = in.Values
	} else {
		conf = &Conf{Values: in.Values}
	}
	return setConf(scs, key, conf)
}

func setConf(scs *state.ContractState, key []byte, conf *Conf) error {
	setKey := genKey(key)
	if err := conf.Validate(setKey); err != nil {
		return err
	}
	return scs.SetData(append(confPrefix, setKey...), serializeConf(conf))
}

func serializeConf(c *Conf) []byte {
	var ret []byte
	if c.On {
		ret = append(ret, 1)
	} else {
		ret = append(ret, 0)
	}
	for _, v := range c.Values {
		ret = append(ret, '\\')
		ret = append(ret, []byte(v)...)
	}
	return ret
}

func deserializeConf(data []byte) *Conf {
	ret := &Conf{
		On:     false,
		Values: strings.Split(string(data), "\\")[1:],
	}
	if data[0] == 1 {
		ret.On = true
	}
	return ret
}

func genKey(key []byte) []byte {
	return []byte(strings.ToUpper(string(key)))
}
