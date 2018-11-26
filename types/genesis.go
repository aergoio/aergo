package types

import "github.com/aergoio/aergo/internal/common"

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800
)

var (
	defaultChainID = ChainID{
		Magic:     "AREGO.IO",
		PublicNet: true,
		Consensus: "dpos",
	}
)

// ChainID represents the identity of the chain.
type ChainID struct {
	Magic     string `json:"magic"`
	PublicNet bool   `json:"public"`
	Consensus string `json:"consensus"`
}

// Genesis represents genesis block
type Genesis struct {
	ID        ChainID           `json:"chain_id,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
	Balance   map[string]*State `json:"alloc"`
	BPs       []string          `json:"bps"`

	// followings are for internal use only
	block     *Block
	voteState *State
}

// Block returns Block corresponding to g.
func (g *Genesis) Block() *Block {
	if g.block == nil {
		g.block = NewBlock(nil, nil, nil, nil, nil, g.Timestamp)
	}
	return g.block
}

// Bytes returns byte-encoded BPs from g.
func (g *Genesis) Bytes() []byte {
	if b, err := common.GobEncode(g); err == nil {
		return b
	}
	return nil
}

// VoteState returns g.voteState
func (g *Genesis) VoteState() *State {
	return g.voteState
}

// SetVoteState sets s to g.VoteState.
func (g *Genesis) SetVoteState(s *State) {
	g.voteState = s
}

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *Genesis {
	return &Genesis{
		ID:        defaultChainID,
		Timestamp: DefaultSeed,
		block:     nil,
	} //TODO embed MAINNET genesis block
}

// GetTestGenesis returns Gensis object for a unit test.
func GetTestGenesis() *Genesis {
	genesis := GetDefaultGenesis()
	genesis.Block()

	return genesis
}

// GetGenesisFromBytes decodes & return Genesis from b.
func GetGenesisFromBytes(b []byte) *Genesis {
	g := &Genesis{}
	if err := common.GobDecode(b, g); err == nil {
		return g
	}
	return nil
}
