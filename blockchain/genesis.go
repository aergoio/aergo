package blockchain

import (
	"github.com/aergoio/aergo/types"
)

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800
)

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *types.Genesis {
	return &types.Genesis{
		Timestamp: DefaultSeed,
		Block:     nil,
	} //TODO embed MAINNET genesis block
}

// GeenesisToBlock returns *types.Block created based on Genesis Information
func GenesisToBlock(gb *types.Genesis) *types.Block {
	genesisBlock := types.NewBlock(nil, nil, 0)
	genesisBlock.Header.Timestamp = gb.Timestamp
	gb.Block = genesisBlock
	return genesisBlock
}
