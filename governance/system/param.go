package system

import (
	"math/big"
	"strings"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

const (
	RESET = -1
)

func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}

//go:generate stringer -type=sysParamIndex
type SysParamIndex int

const (
	BpCount SysParamIndex = iota // BP count
	StakingMin
	GasPrice
	NamePrice
	SysParamMax
)

func (i SysParamIndex) ID() string {
	return strings.ToUpper(i.String())
}

func (i SysParamIndex) Key() []byte {
	return GenProposalKey(i.String())
}

func GetVotingIssues() []types.VotingIssue {
	vi := make([]types.VotingIssue, SysParamMax)
	for i := SysParamIndex(0); i < SysParamMax; i++ {
		vi[int(i)] = i
	}
	return vi
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

func (p *Parameters) Copy() *Parameters {
	param := &Parameters{
		params: map[string]*big.Int{},
	}
	for k, v := range p.params {
		param.params[k] = new(big.Int).Set(v)
	}
	return param
}

func (p *Parameters) SetBpCount(bpCount *big.Int) *big.Int {
	return p.setLastParam(BpCount.ID(), bpCount)
}

func (p *Parameters) GetBpCount() *big.Int {
	return p.getLastParam(BpCount.ID())
}

func (p *Parameters) SetStakingMinimum(stakingMinimum *big.Int) *big.Int {
	return p.setLastParam(StakingMin.ID(), stakingMinimum)
}

func (p *Parameters) GetStakingMinimum() *big.Int {
	return p.getLastParam(StakingMin.ID())
}

func (p *Parameters) SetGasPrice(gasPrice *big.Int) *big.Int {
	return p.setLastParam(GasPrice.ID(), gasPrice)
}

func (p *Parameters) GetGasPrice() *big.Int {
	return p.getLastParam(GasPrice.ID())
}

func (p *Parameters) SetNamePrice(namePrice *big.Int) *big.Int {
	return p.setLastParam(NamePrice.ID(), namePrice)
}

func (p *Parameters) GetNamePrice() *big.Int {
	return p.getLastParam(NamePrice.ID())
}

func (p *Parameters) getLastParam(proposalID string) *big.Int {
	return p.params[proposalID]
}

func (p *Parameters) setLastParam(proposalID string, value *big.Int) *big.Int {
	p.params[proposalID] = value
	return value
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
