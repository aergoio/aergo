package evm

import (
	"errors"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

var (
	logger = log.NewLogger("evm")
)

type EVM struct {
	readonly  bool
	ethState  *ethdb.StateDB
	stateRoot common.Hash
}

func NewEVM(prevStateRoot []byte, ethState *ethdb.StateDB) *EVM {
	return &EVM{
		readonly:  false,
		stateRoot: common.BytesToHash(prevStateRoot),
		ethState:  ethState,
	}
}

func NewEVMCall(queryStateRoot []byte, ethState *ethdb.StateDB) *EVM {
	return &EVM{
		readonly:  true,
		stateRoot: common.BytesToHash(queryStateRoot),
		ethState:  ethState,
	}
}

func (evm *EVM) Query(address []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	queryState := evm.ethState.Copy()
	runtimeCfg := &runtime.Config{
		State:     queryState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	ethOriginAddress := common.BytesToAddress(address)
	contractEthAddress := common.BytesToAddress(contractAddress)
	runtimeCfg.Origin = ethOriginAddress
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := runtime.Call(contractEthAddress, payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (evm *EVM) Call(address common.Address, contract, payload []byte) ([]byte, uint64, error) {
	if evm.readonly {
		return nil, 0, errors.New("cannot call on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     evm.ethState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = address
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := runtime.Call(common.BytesToAddress(contract), payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (evm *EVM) Create(ethAddress common.Address, payload []byte) ([]byte, []byte, uint64, error) {
	if evm.readonly {
		return nil, nil, 0, errors.New("cannot create on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{}

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     evm.ethState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = ethAddress

	ret, ethContractAddress, _, err := runtime.Create(payload, runtimeCfg)
	if err != nil {
		return nil, nil, 0, err
	}

	return ret, ethContractAddress.Bytes(), 0, nil
}
