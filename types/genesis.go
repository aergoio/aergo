package types

import (
	"bytes"

	"github.com/aergoio/aergo/internal/common"
)

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800
)

var (
	defaultChainID = ChainID{
		Magic:     "AREGO.IO",
		PublicNet: true,
		MainNet:   false,
		Consensus: "dpos",
	}
)

const (
	// MagicMax is the max size of the Magic field of ChainID.
	MagicMax = 10
	// ConsensusIDMax is the max size of the Consensus field of ChainID.
	ConsensusIDMax = 10
)

// ChainID represents the identity of the chain.
type ChainID struct {
	PublicNet bool   `json:"public"`
	MainNet   bool   `json:"mainnet"`
	Magic     string `json:"magic"`
	Consensus string `json:"consensus"`
}

// Bytes returns the binary representation of g.ID.
func (cid *ChainID) Bytes() []byte {
	if b, err := common.GobEncode(cid); err == nil {
		return b
	}
	return nil
}

// Equals reports wheter cid equals rhs or not.
func (cid *ChainID) Equals(rhs *ChainID) bool {
	return bytes.Compare(cid.Bytes(), rhs.Bytes()) == 0
}

// Genesis represents genesis block
type Genesis struct {
	ID        ChainID           `json:"chain_id,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
	Balance   map[string]string `json:"balance"`
	BPs       []string          `json:"bps"`

	// followings are for internal use only
	block *Block
}

// Block returns Block corresponding to g.
func (g *Genesis) Block() *Block {
	if g.block == nil {
		g.block = NewBlock(nil, nil, nil, nil, nil, g.Timestamp)
	}
	return g.block
}

// ChainID returns the binary representation of g.ID.
func (g *Genesis) ChainID() []byte {
	return g.ID.Bytes()
}

// Bytes returns byte-encoded BPs from g.
func (g Genesis) Bytes() []byte {
	// Omit the Balance to reduce the resulting data size.
	g.Balance = nil
	if b, err := common.GobEncode(g); err == nil {
		return b
	}
	return nil
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
