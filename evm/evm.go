package evm

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var (
	logger = log.NewLogger("evm")
)

type EVM struct {
	accounts map[common.Address]*state.AccountState
	blocks   map[uint64][]byte

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

func NewEVMQuery(queryStateRoot []byte, ethState *ethdb.StateDB) *EVM {
	return &EVM{
		readonly:  true,
		stateRoot: common.BytesToHash(queryStateRoot),
		ethState:  ethState,
	}
}

func (e *EVM) Query(address []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	queryState := e.ethState.Copy()
	runtimeCfg := &Config{
		State:     queryState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	ethOriginAddress := common.BytesToAddress(address)
	contractEthAddress := common.BytesToAddress(contractAddress)
	runtimeCfg.Origin = ethOriginAddress
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := e.call(contractEthAddress, payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (e *EVM) Call(address common.Address, contract, payload []byte) ([]byte, uint64, error) {
	if e.readonly {
		return nil, 0, errors.New("cannot call on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{
		NoBaseFee: true,
	}

	// create call cfg
	runtimeCfg := &Config{
		State:     e.ethState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = address
	runtimeCfg.GasLimit = 1000000

	ret, gas, err := e.call(common.BytesToAddress(contract), payload, runtimeCfg)
	if err != nil {
		return ret, gas, err
	}

	return ret, gas, nil
}

func (e *EVM) Create(ethAddress common.Address, payload []byte) ([]byte, []byte, uint64, error) {
	if e.readonly {
		return nil, nil, 0, errors.New("cannot create on readonly")
	}

	// create evmCfg
	evmCfg := vm.Config{}

	// create call cfg
	runtimeCfg := &Config{
		State:     e.ethState.GetStateDB(),
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = ethAddress

	ret, ethContractAddress, _, err := e.create(payload, runtimeCfg)
	if err != nil {
		return nil, nil, 0, err
	}

	return ret, ethContractAddress.Bytes(), 0, nil
}

func (e *EVM) GetHashFn() vm.GetHashFunc {
	return func(n uint64) common.Hash {
		blockHash := e.blocks[n]
		return common.BytesToHash(blockHash)
	}
}

func (e *EVM) TransferFn() vm.TransferFunc {
	return func(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
		if senderState := e.accounts[sender]; senderState != nil {
			senderState.SubBalance(amount)
		} else {
			// TODO - get from state
		}
		if receipientState := e.accounts[recipient]; receipientState != nil {
			receipientState.AddBalance(amount)
		} else {
			// TODO - get from state
		}
	}
}

func (e *EVM) CanTransferFn() vm.CanTransferFunc {
	return func(sdb vm.StateDB, addr common.Address, amount *big.Int) bool {
		if state := e.accounts[addr]; state != nil {
			return state.Balance().Cmp(amount) >= 0
		} else {
			// TODO - get from state
		}
		return false
	}
}
