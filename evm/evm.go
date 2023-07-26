package evm

import (
	"github.com/aergoio/aergo-lib/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/ethdb"
)

const (
	rootHashKey  = "roothashkey"
	nullRootHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

var (
	logger = log.NewLogger("evm")
)

type EVM struct {
	levelDB        ethdb.Database
	stateDB        state.Database
	ethState       *state.StateDB
	queryStateRoot common.Hash
	prevStateRoot  common.Hash
}

func NewEVM() *EVM {
	evm := &EVM{}
	return evm
}

func (evm *EVM) LoadDatabase(dbPath string) {
	logger.Info().Msgf("opening a new levelDB for EVM")
	evm.openDatabase(dbPath)

	// set up state
	evm.prevStateRoot = common.Hash{} // FIXME: fetch prev root state hash
	item, err := evm.levelDB.Get([]byte(rootHashKey))
	if err != nil && item == nil {
		// start with null root
		logger.Info().Msg("loaded with null root")
	} else {
		evm.prevStateRoot.SetBytes(item)
	}
	evm.ethState, _ = state.New(evm.prevStateRoot, evm.stateDB, nil)
	if evm.ethState == nil {
		logger.Error().Msgf("eth state not created")
	}
	logger.Info().Msgf("created eth state with root %s", evm.prevStateRoot.String())

}

func (evm *EVM) openDatabase(dbPath string) error {
	evm.levelDB, _ = rawdb.NewLevelDBDatabase(dbPath, 128, 1024, "", false)
	evm.stateDB = state.NewDatabase(evm.levelDB)
	return nil
}

func (evm *EVM) CloseDatabase() {
	logger.Info().Msgf("closing levelDB for EVM with root %s", evm.prevStateRoot.String())
	evm.levelDB.Close()
}

func (evm *EVM) Commit() error {
	evm.queryStateRoot = evm.prevStateRoot
	evm.prevStateRoot, _ = evm.ethState.Commit(true)
	evm.levelDB.Put([]byte(rootHashKey), evm.prevStateRoot.Bytes())
	evm.ethState, _ = state.New(evm.prevStateRoot, evm.stateDB, nil)
	logger.Info().Msgf("commiting eth state with root hash %s", evm.prevStateRoot.String())
	return nil
}

func (evm *EVM) Query(originAddress []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		Debug:     false,
		NoBaseFee: true,
	}

	// create call cfg
	queryState, _ := state.New(evm.queryStateRoot, evm.stateDB, nil)
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
	// create evmCfg
	evmCfg := vm.Config{
		Debug:     false,
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
	// create evmCfg
	evmCfg := vm.Config{
		Debug: false,
	}

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
