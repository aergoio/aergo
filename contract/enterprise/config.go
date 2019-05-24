package enterprise

import (
	"strings"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var confPrefix = []byte("conf\\")

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
	data, err := scs.GetData(append(confPrefix, key...))
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
	return scs.SetData(append(confPrefix, key...), serializeConf(conf))
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
