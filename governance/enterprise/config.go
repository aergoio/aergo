package enterprise

import (
	"fmt"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var confPrefix = []byte("conf\\")

const (
	RPCPermissions = "RPCPERMISSIONS"
	P2PWhite       = "P2PWHITE"
	P2PBlack       = "P2PBLACK"
	AccountWhite   = "ACCOUNTWHITE"
)

// EnterpriseKeyDict is represent allowed key list and used when validate tx, int values are meaningless.
var enterpriseKeyDict = map[string]int{
	RPCPermissions: 1,
	P2PWhite:       2,
	P2PBlack:       3,
	AccountWhite:   4,
}

type Conf struct {
	On     bool
	Values []string
}

func (c *Conf) AppendValue(r string) {
	if c.Values == nil {
		c.Values = []string{}
	}
	c.Values = append(c.Values, r)
}
func (c *Conf) RemoveValue(r string) {
	for i, v := range c.Values {
		if v == r {
			c.Values = append(c.Values[:i], c.Values[i+1:]...)
			break
		}
	}
}

func (c *Conf) Validate(key []byte, context *EnterpriseContext) error {
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
	case AccountWhite:
		for _, v := range context.Admins {
			address := types.EncodeAddress(v)
			if context.HasConfValue(address) {
				return nil
			}
		}
		return fmt.Errorf("the values of %s should have at least one admin address", strKey)
	default:
		return nil
	}
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
	ret := &types.EnterpriseConfig{Key: key}
	if strings.ToUpper(key) == "PERMISSIONS" {
		for k := range enterpriseKeyDict {
			ret.Values = append(ret.Values, k)
		}
	} else {
		conf, err := getConf(scs, []byte(key))
		if err != nil {
			return nil, err
		}
		if conf != nil {
			ret.On = conf.On
			ret.Values = conf.Values
		}
	}
	return ret, nil
}

func enableConf(scs *state.ContractState, key []byte, value bool) (*Conf, error) {
	conf, err := getConf(scs, key)
	if err != nil {
		return nil, err
	}
	if conf != nil {
		conf.On = value
	} else {
		conf = &Conf{On: value}
	}
	return conf, nil
}

func getConf(scs *state.ContractState, key []byte) (*Conf, error) {
	data, err := scs.GetData(append(confPrefix, genKey(key)...))
	if err != nil || data == nil {
		return nil, err
	}
	return deserializeConf(data), err
}

func setConfValues(scs *state.ContractState, key []byte, values []string) (*Conf, error) {
	conf, err := getConf(scs, key)
	if err != nil {
		return nil, err
	}
	if conf != nil {
		conf.Values = values
	} else {
		conf = &Conf{Values: values}
	}
	return conf, nil
}

func setConf(scs *state.ContractState, key []byte, conf *Conf) error {
	return scs.SetData(append(confPrefix, genKey(key)...), serializeConf(conf))
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
