package types

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/aergoio/aergo/internal/common"
)

const (
	// DefaultSeed is temporary const to create same genesis block with no configuration
	DefaultSeed = 1530838800

	blockVersionNil = math.MinInt32
)

var (
	nilChainID = ChainID{
		Version:   blockVersionNil,
		Magic:     "",
		PublicNet: false,
		MainNet:   false,
		Consensus: "",
	}

	defaultChainID = ChainID{
		Version:   0,
		Magic:     "AREGO.IO",
		PublicNet: false,
		MainNet:   false,
		Consensus: "sbp",
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
	Version   int32  `json:"-"`
	PublicNet bool   `json:"public"`
	MainNet   bool   `json:"mainnet"`
	Magic     string `json:"magic"`
	Consensus string `json:"consensus"`
}

// NewChainID returns a new ChainID initialized as nilChainID.
func NewChainID() *ChainID {
	nilCID := nilChainID

	return &nilCID
}

// Bytes returns the binary representation of cid.
func (cid *ChainID) Bytes() []byte {
	var w bytes.Buffer
	if err := binary.Write(&w, binary.LittleEndian, *cid); err != nil {
		return nil
	}
	return w.Bytes()
}

// Read deserialize data as a ChainID.
func (cid *ChainID) Read(data []byte) error {
	return binary.Read(bytes.NewBuffer(data), binary.LittleEndian, cid)
}

// AsDefault set *cid to the default chaind id (cid must be a valid pointer).
func (cid *ChainID) AsDefault() {
	*cid = defaultChainID
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

// Consensus retruns g.ID.Consensus.
func (g Genesis) Consensus() string {
	return g.ID.Consensus
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
