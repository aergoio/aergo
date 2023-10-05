package system

import (
	"math/big"
	"strings"
	"sync"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

type parameters struct {
	mutex  sync.RWMutex
	params map[string]*big.Int
}

func (p *parameters) setParam(proposalID string, value *big.Int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.params[proposalID] = value
}

// save the new value for the param, to be active on the next block
func (p *parameters) setNextParam(proposalID string, value *big.Int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.params[nextParamKey(proposalID)] = value
}

func (p *parameters) delNextParam(proposalID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.params, nextParamKey(proposalID))
}

func (p *parameters) getNextParam(proposalID string) *big.Int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.params[nextParamKey(proposalID)]
}

func (p *parameters) getParam(proposalID string) *big.Int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.params[proposalID]
}

func nextParamKey(id string) string {
	return id + "next"
}

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
	systemParams *parameters = &parameters{
		mutex:  sync.RWMutex{},
		params: map[string]*big.Int{},
	}

	//DefaultParams is for aergo v1 compatibility
	DefaultParams = map[string]*big.Int{
		stakingMin.ID(): types.StakingMinimum,
		gasPrice.ID():   types.NewAmount(50, types.Gaer), // 50 gaer
		namePrice.ID():  types.NewAmount(1, types.Aergo), // 1 aergo
	}
)

func genParamKey(id string) []byte {
	return []byte("param\\" + strings.ToUpper(id))
}

// This is also called on chain reorganization
func InitSystemParams(g dataGetter, bpCount int) {
	// discard any new params computed for the next block
	CommitParams(false)
	// (re)load param values from database
	initDefaultBpCount(bpCount)
	systemParams = loadParams(g)
}

// This function must be called before all the aergosvr
// services start.
func initDefaultBpCount(count int) {
	// Ensure that it is not modified after it is initialized.
	if DefaultParams[bpCount.ID()] == nil {
		DefaultParams[bpCount.ID()] = big.NewInt(int64(count))
	}
}

// load the params from the database or use the default values
func loadParams(g dataGetter) *parameters {
	ret := map[string]*big.Int{}
	for i := sysParamIndex(0); i < sysParamMax; i++ {
		id := i.ID()
		data, err := g.GetData(genParamKey(id))
		if err != nil {
			panic("could not load blockchain parameter")
		}
		if data != nil {
			ret[id] = new(big.Int).SetBytes(data)
		} else {
			ret[id] = DefaultParams[id]
		}
	}
	return &parameters{
		mutex:  sync.RWMutex{},
		params: ret,
	}
}

func updateParam(s dataSetter, id string, value *big.Int) error {
	// save the param to the database (in a db txn, commit when the block is connected)
	if err := s.SetData(genParamKey(id), value.Bytes()); err != nil {
		return err
	}
	// save the new value for the param, only active on the next block
	systemParams.setNextParam(id, value)
	return nil
}

// if a system param was changed, apply or discard its new value
func CommitParams(apply bool) {
	for i := sysParamIndex(0); i < sysParamMax; i++ {
		id := i.ID()
		// check if the param has a new value
		if param := systemParams.getNextParam(id); param != nil {
			if apply {
				// set the new value for the current block
				systemParams.setParam(id, param)
			}
			// delete the new value
			systemParams.delNextParam(id)
		}
	}
}

// get the param value for the next block
func GetNextParam(proposalID string) *big.Int {
	// check the value for the next block
	if val := systemParams.getNextParam(proposalID); val != nil {
		return val
	}
	// check the value for the current block
	if val := systemParams.getParam(proposalID); val != nil {
		return val
	}
	// default value
	return DefaultParams[proposalID]
}

// get the param value for the current block
func GetParam(proposalID string) *big.Int {
	if val := systemParams.getParam(proposalID); val != nil {
		return val
	}
	return DefaultParams[proposalID]
}

// these 4 functions are reading the param value for the current block

func GetStakingMinimum() *big.Int {
	return GetParam(stakingMin.ID())
}

func GetGasPrice() *big.Int {
	return GetParam(gasPrice.ID())
}

func GetNamePrice() *big.Int {
	return GetParam(namePrice.ID())
}

func GetBpCount() int {
	return int(GetParam(bpCount.ID()).Uint64())
}

// these functions are reading the param value directly from the state

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
