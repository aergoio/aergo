package types

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800
)

// Genesis represents genesis block
type Genesis struct {
	Timestamp int64             `json:"timestamp,omitempty"`
	Balance   map[string]*State `json:"alloc"`
	BPIds     []string          `json:"bpids"`

	// followings are for internal use only
	Block     *Block `json:"-"`
	VoteState *State `json:"-"`
}

// GetBlock returns Block corresponding to g.
func (g *Genesis) GetBlock() *Block {
	if g.Block == nil {
		g.Block = NewBlock(nil, nil, nil, nil, nil, g.Timestamp)
	}
	return g.Block
}

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *Genesis {
	return &Genesis{
		Timestamp: DefaultSeed,
		Block:     nil,
	} //TODO embed MAINNET genesis block
}

// GetTestGenesis returns Gensis object for a unit test.
func GetTestGenesis() *Genesis {
	genesis := GetDefaultGenesis()
	genesis.GetBlock()

	return genesis
}
