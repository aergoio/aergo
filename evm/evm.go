package evm

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/ethdb"
	"github.com/aergoio/aergo/v2/state/statedb"
	"github.com/aergoio/aergo/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var (
	logger = log.NewLogger("evm")
)

type ChainAccessor interface {
	GetBlockByNo(blockNo types.BlockNo) (*types.Block, error)
	GetBestBlock() (*types.Block, error)
}

type EVM struct {
	readonly      bool
	chainAccessor ChainAccessor
	luaState      *statedb.StateDB
	ethState      *ethdb.StateDB
	stateRoot     common.Hash
}

func NewEVM(prevStateRoot []byte, chainAccessor ChainAccessor, luaState *statedb.StateDB, ethState *ethdb.StateDB) *EVM {
	return &EVM{
		readonly:      false,
		chainAccessor: chainAccessor,
		stateRoot:     common.BytesToHash(prevStateRoot),
		luaState:      luaState,
		ethState:      ethState,
	}
}

func NewEVMQuery(chainAccessor ChainAccessor, queryStateRoot []byte, luaState *statedb.StateDB, ethState *ethdb.StateDB) *EVM {
	return &EVM{
		readonly:  true,
		luaState:  nil,
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
		block, err := e.chainAccessor.GetBlockByNo(n)
		if err != nil {
			return common.Hash{}
		}
		return common.BytesToHash(block.Hash)
	}
}

func (e *EVM) TransferFn() vm.TransferFunc {
	return func(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
		senderAccState, err := state.GetAccountState(e.ethState.GetId(sender), e.luaState, e.ethState)
		if err != nil {
			panic("impossible") // FIXME
		}
		receipientAccState, err := state.GetAccountState(e.ethState.GetId(recipient), e.luaState, e.ethState)
		if err != nil {
			panic("impossible") // FIXME
		}
		err = state.SendBalance(senderAccState, receipientAccState, amount)
		if err != nil {
			panic("impossible") // FIXME
		}
	}
}

func (e *EVM) CanTransferFn() vm.CanTransferFunc {
	return func(sdb vm.StateDB, addr common.Address, amount *big.Int) bool {
		accState, err := state.GetAccountState(e.ethState.GetId(addr), e.luaState, e.ethState)
		if err != nil {
			panic("impossible") // FIXME
		}
		return accState.Balance().Cmp(amount) >= 0
	}
}
