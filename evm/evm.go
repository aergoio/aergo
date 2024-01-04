package evm

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/fee"
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

	bs  *state.BlockState
	sdb *StateDB
}

func NewEVM(chainAccessor ChainAccessor, blockState *state.BlockState) *EVM {
	return &EVM{
		readonly:      false,
		chainAccessor: chainAccessor,
		bs:            blockState,
		sdb:           NewStateDB(blockState),
	}
}

func NewEVMQuery(chainAccessor ChainAccessor, blockState *state.BlockState) *EVM {
	return &EVM{
		readonly:      true,
		chainAccessor: chainAccessor,
		bs:            blockState,
	}
}

func (e *EVM) Query(address []byte, contractAddress []byte, payload []byte) ([]byte, *big.Int, error) {

	// create evmCfg
	ethOriginAddress := common.BytesToAddress(address)
	contractEthAddress := common.BytesToAddress(contractAddress)
	cfg := NewConfig(
		e.bs.Block().ChainID,
		ethOriginAddress,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		0,
		e.bs.GasPrice(),
		types.NewZeroAmount(),
		e.sdb,
	)

	// create call cfg
	ret, leftOverGas, err := e.call(contractEthAddress, payload, cfg)
	gasUsed := cfg.GasLimit - leftOverGas
	feeUsed := fee.CalcFee(cfg.GasPrice, gasUsed)

	if err != nil {
		return ret, feeUsed, err
	}
	return ret, feeUsed, nil
}

func (e *EVM) Call(address, contract common.Address, payload []byte, amount *big.Int, gasLimit uint64) ([]byte, *big.Int, error) {
	if e.readonly {
		return nil, nil, errors.New("cannot call on readonly")
	}

	// create evmCfg
	cfg := NewConfig(
		e.bs.Block().ChainID,
		address,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		gasLimit,
		e.bs.GasPrice(),
		amount,
		e.sdb,
	)

	ret, leftOverGas, err := e.call(contract, payload, cfg)
	gasUsed := cfg.GasLimit - leftOverGas
	feeUsed := fee.CalcFee(cfg.GasPrice, gasUsed)

	if err != nil {
		return ret, feeUsed, err
	}
	return ret, feeUsed, nil
}

func (e *EVM) Create(sender common.Address, payload []byte, gasLimit uint64) ([]byte, []byte, *big.Int, error) {
	if e.readonly {
		return nil, nil, nil, errors.New("cannot create on readonly")
	}

	// create evmCfg
	cfg := NewConfig(
		e.bs.Block().ChainID,
		sender,
		ethdb.GetAddressEth(e.bs.Block().CoinbaseAccount),
		e.bs.Block().BlockNo,
		uint64(e.bs.Block().Timestamp),
		gasLimit,
		e.bs.GasPrice(),
		types.NewZeroAmount(),
		e.sdb,
	)

	ret, ethContractAddress, leftOverGas, err := e.create(payload, cfg)
	gasUsed := cfg.GasLimit - leftOverGas
	feeUsed := fee.CalcFee(cfg.GasPrice, gasUsed)
	if err != nil {
		return nil, nil, feeUsed, err
	}

	return ret, ethContractAddress.Bytes(), feeUsed, nil
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
		db.SubBalance(sender, amount)
		db.AddBalance(recipient, amount)
	}
}

func (e *EVM) CanTransferFn() vm.CanTransferFunc {
	return func(sdb vm.StateDB, addr common.Address, amount *big.Int) bool {
		return sdb.GetBalance(addr).Cmp(amount) >= 0
	}
}
