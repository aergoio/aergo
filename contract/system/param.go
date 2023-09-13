package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type Parameters map[string]*big.Int

const (
	RESET = -1
)

//go:generate stringer -type=sysParamIndex
type SysParamIndex int

const (
	BpCount SysParamIndex = iota // BP count
	StakingMin
	GasPrice
	NamePrice
	SysParamMax
)

var (
	systemParams Parameters

	//DefaultParams is for aergo v1 compatibility
	DefaultParams = map[string]*big.Int{
		StakingMin.ID(): types.StakingMinimum,
		GasPrice.ID():   types.NewAmount(50, types.Gaer), // 50 gaer
		NamePrice.ID():  types.NewAmount(1, types.Aergo), // 1 aergo
	}
)

func InitSystemParams(g DataGetter, bpCount int) {
	initDefaultBpCount(bpCount)
	systemParams = loadParam(g)
}

func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}

func loadParam(g DataGetter) Parameters {
	ret := map[string]*big.Int{}
	for i := SysParamIndex(0); i < SysParamMax; i++ {
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

func (p Parameters) getLastParam(proposalID string) *big.Int {
	if val, ok := p[proposalID]; ok {
		return val
	}
	return DefaultParams[proposalID]
}

func (p Parameters) setLastParam(proposalID string, value *big.Int) *big.Int {
	p[proposalID] = value
	return value
}

func updateParam(s DataSetter, id string, value *big.Int) (*big.Int, error) {
	if err := s.SetData(genParamKey(id), value.Bytes()); err != nil {
		return nil, err
	}
	ret := systemParams.setLastParam(id, value)
	return ret, nil
}

func GetStakingMinimum() *big.Int {
	return GetParam(StakingMin.ID())
}

func GetGasPrice() *big.Int {
	return GetParam(GasPrice.ID())
}

func GetNamePrice() *big.Int {
	return GetParam(NamePrice.ID())
}

func GetNamePriceFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, NamePrice)
}

func GetStakingMinimumFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, StakingMin)
}

func GetGasPriceFromState(ar AccountStateReader) *big.Int {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		panic("could not open system state when get gas price")
	}
	return getParamFromState(scs, GasPrice)
}

func getParamFromState(scs *state.ContractState, id SysParamIndex) *big.Int {
	data, err := scs.GetInitialData(genParamKey(id.ID()))
	if err != nil {
		panic("could not get blockchain parameter")
	}
	if data == nil {
		return DefaultParams[id.ID()]
	}
	return new(big.Int).SetBytes(data)
}
