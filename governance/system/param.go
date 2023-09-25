package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/state"
)

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

func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}

// func LoadParam(g DataGetter) Parameters {
// 	ret := map[string]*big.Int{}
// 	for i := SysParamIndex(0); i < SysParamMax; i++ {
// 		id := i.ID()
// 		data, err := g.GetData(genParamKey(id))
// 		if err != nil {
// 			panic("could not load blockchain parameter")
// 		}
// 		if data == nil {
// 			ret[id] = DefaultParams[id]
// 			continue
// 		}
// 		ret[id] = new(big.Int).SetBytes(data)
// 	}
// 	return ret
// }

// params in memory
type Parameters struct {
	params map[string]*big.Int
}

func NewParameters() *Parameters {
	return &Parameters{
		params: map[string]*big.Int{},
	}
}

func (p *Parameters) getLastParam(proposalID string) *big.Int {
	return p.params[proposalID]
}

func (p *Parameters) setLastParam(proposalID string, value *big.Int) *big.Int {
	p.params[proposalID] = value
	return value
}

func (p *Parameters) UpdateParam(s DataSetter, id string, value *big.Int) (*big.Int, error) {
	if err := s.SetData(genParamKey(id), value.Bytes()); err != nil {
		return nil, err
	}
	ret := p.setLastParam(id, value)
	return ret, nil
}

func (p *Parameters) GetBpCount() *big.Int {
	return p.getLastParam(BpCount.ID())
}

func (p *Parameters) GetStakingMinimum() *big.Int {
	return p.getLastParam(StakingMin.ID())
}

func (p *Parameters) GetGasPrice() *big.Int {
	return p.getLastParam(GasPrice.ID())
}

func (p *Parameters) GetNamePrice() *big.Int {
	return p.getLastParam(NamePrice.ID())
}

// params in state
func GetBpCountFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, BpCount)
}

func GetNamePriceFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, NamePrice)
}

func GetStakingMinimumFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, StakingMin)
}

func GetGasPriceFromAccountState(ar AccountStateReader) *big.Int {
	scs, err := ar.GetSystemAccountState()
	if err != nil {
		panic("could not open system state when get gas price")
	}
	return GetGasPriceFromState(scs)
}

func GetGasPriceFromState(scs *state.ContractState) *big.Int {
	return getParamFromState(scs, GasPrice)
}

func getParamFromState(scs *state.ContractState, id SysParamIndex) *big.Int {
	data, err := scs.GetInitialData(genParamKey(id.ID()))
	if err != nil {
		panic("could not get blockchain parameter")
	}
	if data == nil {
		return nil
	}
	return new(big.Int).SetBytes(data)
}

/*
func updateParam(s DataSetter, id string, value *big.Int) (*big.Int, error) {
	if err := s.SetData(genParamKey(id), value.Bytes()); err != nil {
		return nil, err
	}
	ret := systemParams.setLastParam(id, value)
	return ret, nil
}
*/
