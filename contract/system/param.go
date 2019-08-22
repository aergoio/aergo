package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/types"
)

type parameters map[string]*big.Int

const (
	BPCOUNT    = "BPCOUNT"
	STAKINGMIN = "STAKINGMIN"
)

var (
	systemParams parameters

	//DefaultParams is for aergo v1 compatibility
	DefaultParams   = map[string]*big.Int{STAKINGMIN: types.StakingMinimum}
	systemParamList = []string{BPCOUNT, STAKINGMIN}
)

func InitSystemParams(g dataGetter, bpCount int) {
	InitDefaultBpCount(bpCount)
	systemParams = loadParam(g)
}
func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}
func loadParam(g dataGetter) parameters {
	ret := map[string]*big.Int{}
	for _, id := range systemParamList {
		data, err := g.GetData(genParamKey(id))
		if err != nil {
			panic("could not load blockchain parameter")
		}
		if data == nil {
			continue
		}
		_ = systemParams.setLastParam(id, new(big.Int).SetBytes(data))
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
