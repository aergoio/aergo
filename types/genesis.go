package types

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800
)

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *Genesis {
	return &Genesis{
		Timestamp: DefaultSeed,
		Block:     nil,
	} //TODO embed MAINNET genesis block
}

// GenesisToBlock returns *types.Block created based on Genesis Information
func GenesisToBlock(gb *Genesis) *Block {
	genesisBlock := NewBlock(nil, nil, 0)
	genesisBlock.Header.Timestamp = gb.Timestamp
	gb.Block = genesisBlock
	return genesisBlock
}

// GetTestGenesis returns Gensis object for a unit test.
func GetTestGenesis() *Genesis {
	genesis := GetDefaultGenesis()
	GenesisToBlock(genesis)

	return genesis
}
