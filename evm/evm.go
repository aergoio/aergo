package evm

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/state/ethdb"
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

	txContext *types.Tx
	bs        *state.BlockState
}

func NewEVM(chainAccessor ChainAccessor, txContext *types.Tx, blockState *state.BlockState) *EVM {
	return &EVM{
		readonly:      false,
		txContext:     txContext,
		chainAccessor: chainAccessor,
		bs:            blockState,
	}
}

func NewEVMQuery(chainAccessor ChainAccessor, queryStateRoot []byte, blockState *state.BlockState) *EVM {
	return &EVM{
		readonly:      true,
		chainAccessor: chainAccessor,
		bs:            blockState,
	}
}

func (e *EVM) Query(address []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg

	ethOriginAddress := common.BytesToAddress(address)
	contractEthAddress := common.BytesToAddress(contractAddress)
	queryState := e.bs.EvmStateDB.GetStateDB().Copy()
	evmCfg := NewConfig(
		e.bs.Block().ChainID,
		ethOriginAddress,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		1000000, e.bs.GasPrice(), big.NewInt(0), queryState,
	)

	// create call cfg
	ret, leftOverGas, err := e.call(contractEthAddress, payload, evmCfg)
	gasUsed := evmCfg.GasLimit - leftOverGas
	if err != nil {
		return ret, gasUsed, err
	}

	return ret, gasUsed, nil
}

func (e *EVM) Call(address common.Address, contract, payload []byte) ([]byte, uint64, error) {
	if e.readonly {
		return nil, 0, errors.New("cannot call on readonly")
	}

	// create evmCfg
	contractEth := common.BytesToAddress(contract)
	queryState := e.bs.EvmStateDB.GetStateDB().Copy()
	cfg := NewConfig(
		e.bs.Block().ChainID,
		address,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		1000000, e.bs.GasPrice(), big.NewInt(0), queryState,
	)

	ret, leftOverGas, err := e.call(contractEth, payload, cfg)
	gasUsed := cfg.GasLimit - leftOverGas
	if err != nil {
		return ret, gasUsed, err
	}

	return ret, gasUsed, nil
}

func (e *EVM) Create(sender common.Address, payload []byte) ([]byte, []byte, uint64, error) {
	if e.readonly {
		return nil, nil, 0, errors.New("cannot create on readonly")
	}

	// create evmCfg
	queryState := e.bs.EvmStateDB.GetStateDB().Copy()
	cfg := NewConfig(
		e.bs.Block().ChainID,
		sender,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		1000000, e.bs.GasPrice(), big.NewInt(0), queryState,
	)

	ret, ethContractAddress, leftOverGas, err := e.create(payload, cfg)
	gasUsed := cfg.GasLimit - leftOverGas
	if err != nil {
		return nil, nil, gasUsed, err
	}

	return ret, ethContractAddress.Bytes(), gasUsed, nil
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
		senderId := e.bs.EvmStateDB.GetId(sender)
		senderAccState, err := state.GetAccountState(senderId, e.bs.LuaStateDB, e.bs.EvmStateDB)
		if err != nil {
			panic("impossible") // FIXME
		}
		receipientId := e.bs.EvmStateDB.GetId(sender)
		receipientAccState, err := state.GetAccountState(receipientId, e.bs.LuaStateDB, e.bs.EvmStateDB)
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
		addrId := e.bs.EvmStateDB.GetId(addr)
		accState, err := state.GetAccountState(addrId, e.bs.LuaStateDB, e.bs.EvmStateDB)
		if err != nil {
			panic("impossible") // FIXME
		}
		return accState.Balance().Cmp(amount) >= 0
	}
}
