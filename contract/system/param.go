package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/types"
)

type parameters map[string]*big.Int

const (
	RESET = -1
)

var (
	systemParams parameters

	//DefaultParams is for aergo v1 compatibility
	DefaultParams = map[string]*big.Int{stakingMin.ID(): types.StakingMinimum}
)

func InitSystemParams(g dataGetter, bpCount int) {
	initDefaultBpCount(bpCount)
	systemParams = loadParam(g)
}

func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}

func loadParam(g dataGetter) parameters {
	ret := map[string]*big.Int{}
	for i := sysParamIndex(0); i < sysParamMax; i++ {
		id := i.ID()
		data, err := g.GetData(genParamKey(id))
		if err != nil {
			panic("could not load blockchain parameter")
		}
		if data == nil {
			ret[id] = DefaultParams[id]
			continue
		}
		ret[id] = new(big.Int).SetBytes(data)
	}
	return ret
}

func (p parameters) getLastParam(proposalID string) *big.Int {
	if val, ok := p[proposalID]; ok {
		return val
	}
	return DefaultParams[proposalID]
}

func (p parameters) setLastParam(proposalID string, value *big.Int) *big.Int {
	p[proposalID] = value
	return value
}

func updateParam(s dataSetter, id string, value *big.Int) (*big.Int, error) {
	if err := s.SetData(genParamKey(id), value.Bytes()); err != nil {
		return nil, err
	}
	ret := systemParams.setLastParam(id, value)
	return ret, nil
}
