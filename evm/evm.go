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

var (
	logger   = log.NewLogger("evm")
	levelDB  ethdb.Database
	stateDB  state.Database
	ethState *state.StateDB
)

func init() {
	// load previous state or create a new state
}

func LoadDatabase(dbPath string) {
	logger.Info().Msgf("opening a new levelDB for EVM")
	openDatabase(dbPath)

	// set up state

	stateRoot := common.Hash{} // FIXME: fetch prev root state hash
	ethState, _ = state.New(stateRoot, stateDB, nil)
	if ethState == nil {
		logger.Error().Msgf("eth state not created")
	}
	logger.Info().Msgf("created eth state")

}

func openDatabase(dbPath string) error {
	levelDB, _ = rawdb.NewLevelDBDatabase(dbPath, 128, 1024, "", false)
	stateDB = state.NewDatabase(levelDB)
	return nil
}

func CloseDatabase() {
	logger.Info().Msgf("closing levelDB for EVM")
	levelDB.Close()
}

func Commit() error {
	nextRoot, _ := ethState.Commit(true)
	ethState, _ = state.New(nextRoot, stateDB, nil) // FIXME: save new root state hash

	return nil
}

func Call(originAddress []byte, contractAddress []byte, payload []byte) ([]byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		Debug:     false,
		NoBaseFee: true,
	}

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     ethState,
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

func Create(originAddress []byte, payload []byte) ([]byte, []byte, uint64, error) {
	// create evmCfg
	evmCfg := vm.Config{
		Debug: false,
	}

	ethAddress := common.BytesToAddress(originAddress)

	// create call cfg
	runtimeCfg := &runtime.Config{
		State:     ethState,
		EVMConfig: evmCfg,
	}

	runtimeCfg.Origin = ethAddress

	ret, ethContractAddress, _, err := runtime.Create(payload, runtimeCfg)
	if err != nil {
		return nil, nil, 0, err
	}

	return ret, ethContractAddress.Bytes(), 0, nil
}
