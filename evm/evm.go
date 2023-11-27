package evm

import (
	"errors"

	"github.com/aergoio/aergo-lib/log"
	key "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

var (
	logger = log.NewLogger("evm")
)

type EVM struct {
	readonly  bool
	ethState  *state.StateDB
	stateRoot common.Hash
}

func NewEVM(prevStateRoot []byte, ethState *state.StateDB) *EVM {
	return &EVM{
		readonly:  false,
		stateRoot: common.BytesToHash(prevStateRoot),
		ethState:  ethState,
	}
}

func NewEVMCall(queryStateRoot []byte, ethState *state.StateDB) *EVM {
	return &EVM{
		readonly:  true,
		stateRoot: common.BytesToHash(queryStateRoot),
		ethState:  ethState,
	}
}

func (evm *EVM) Query(originAddress []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	queryState, _ := state.New(evm.stateRoot, evm.ethState.Database(), nil)
	runtimeCfg := &runtime.Config{
		State:     queryState,
		EVMConfig: evmCfg,
	}

	ethOriginAddress := common.BytesToAddress(originAddress)
	contractEthAddress := common.BytesToAddress(contractAddress)
	runtimeCfg.Origin = ethOriginAddress
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := runtime.Call(contractEthAddress, payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (evm *EVM) Call(originAddress []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	if evm.readonly {
		return nil, 0, errors.New("cannot call on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     evm.ethState,
		EVMConfig: evmCfg,
	}

	ethOriginAddress := common.BytesToAddress(originAddress)
	contractEthAddress := common.BytesToAddress(contractAddress)
	runtimeCfg.Origin = ethOriginAddress
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := runtime.Call(contractEthAddress, payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (evm *EVM) Create(originAddress []byte, payload []byte) ([]byte, []byte, uint64, error) {
	if evm.readonly {
		return nil, nil, 0, errors.New("cannot create on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{}

	ethAddress := common.BytesToAddress(originAddress)

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     evm.ethState,
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = ethAddress

	ret, ethContractAddress, _, err := runtime.Create(payload, runtimeCfg)
	if err != nil {
		return nil, nil, 0, err
	}

	return ret, ethContractAddress.Bytes(), 0, nil
}

func ConvertAddress(aergoAddress []byte) []byte {
	if unCompressed := key.ConvAddressUncompressed(aergoAddress); unCompressed != nil {
		return common.BytesToAddress(unCompressed).Bytes()
	}
	return nil
}
