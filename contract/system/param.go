package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type parameters map[string]*big.Int

const (
	RESET = -1
)

//go:generate stringer -type=sysParamIndex
type sysParamIndex int

const (
	bpCount sysParamIndex = iota // BP count
	stakingMin
	gasPrice
	namePrice
	sysParamMax
)

var (
	systemParams parameters

	//DefaultParams is for aergo v1 compatibility
	DefaultParams = map[string]*big.Int{
		stakingMin.ID(): types.StakingMinimum,
		gasPrice.ID():   types.NewAmount(50, types.Gaer), // 50 gaer
		namePrice.ID():  types.NewAmount(1, types.Aergo), // 1 aergo
	}
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

func GetStakingMinimum() *big.Int {
	return GetParam(stakingMin.ID())
}

func GetGasPrice() *big.Int {
	return GetParam(gasPrice.ID())
}

func GetNamePrice() *big.Int {
	return GetParam(namePrice.ID())
}

func GetNamePriceFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, namePrice)
}

func GetStakingMinimumFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, stakingMin)
}

func GetGasPriceFromState(ar AccountStateReader) *big.Int {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		panic("could not open system state when get gas price")
	}
	return getParamFromState(scs, gasPrice)
}

func getParamFromState(scs *state.ContractState, id sysParamIndex) *big.Int {
	data, err := scs.GetInitialData(genParamKey(id.ID()))
	if err != nil {
		panic("could not get blockchain parameter")
	}
	if data == nil {
		return DefaultParams[id.ID()]
	}
	return new(big.Int).SetBytes(data)
}
