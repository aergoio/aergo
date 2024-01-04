package evm

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

type Config struct {
	ChainConfig *params.ChainConfig
	Difficulty  *big.Int
	Origin      common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        uint64
	GasLimit    uint64
	GasPrice    *big.Int
	Value       *big.Int
	Debug       bool
	EVMConfig   vm.Config
	BaseFee     *big.Int
	BlobBaseFee *big.Int
	BlobHashes  []common.Hash
	BlobFeeCap  *big.Int
	Random      *common.Hash

	State *StateDB
}

// TODO : modify runtime config and functions to be compatible with aergo
func NewConfig(
	chainID []byte,
	sender, coinbase common.Address,
	blockNo, time, gasLimit uint64,
	gasPrice, Value *big.Int,
	state *StateDB,
) *Config {
	cfg := new(Config)
	cfg.Origin = sender
	cfg.Coinbase = coinbase
	cfg.BlockNumber = new(big.Int).SetUint64(blockNo)
	cfg.Time = time
	cfg.GasLimit = gasLimit
	cfg.GasPrice = gasPrice
	cfg.Value = Value
	cfg.State = state
	cfg.EVMConfig = vm.Config{
		NoBaseFee: true,
	}
	setDefault(cfg)

	// overwrite chainid
	cfg.ChainConfig.ChainID = new(big.Int).SetBytes(chainID)

	return cfg
}

// sets defaults on the config
func setDefault(cfg *Config) {
	if cfg.ChainConfig == nil {
		cfg.ChainConfig = &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			DAOForkBlock:        new(big.Int),
			DAOForkSupport:      false,
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			MuirGlacierBlock:    new(big.Int),
			BerlinBlock:         new(big.Int),
			LondonBlock:         new(big.Int),
		}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.GasLimit == 0 {
		cfg.GasLimit = math.MaxUint64
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.BaseFee == nil {
		cfg.BaseFee = big.NewInt(params.InitialBaseFee) // FIXME
	}
	if cfg.BlobBaseFee == nil {
		cfg.BlobBaseFee = big.NewInt(params.BlobTxMinBlobGasprice) // FIXME
	}
}

func NewEnv(cfg *Config, evm *EVM) *vm.EVM {
	txContext := vm.TxContext{
		Origin:     cfg.Origin,
		GasPrice:   cfg.GasPrice,
		BlobHashes: cfg.BlobHashes,
		BlobFeeCap: cfg.BlobFeeCap,
	}
	blockContext := vm.BlockContext{
		CanTransfer: evm.CanTransferFn(),
		Transfer:    evm.TransferFn(),
		GetHash:     evm.GetHashFn(),
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
		BaseFee:     cfg.BaseFee,
		BlobBaseFee: cfg.BlobBaseFee,
		Random:      cfg.Random,
	}

	return vm.NewEVM(blockContext, txContext, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}
