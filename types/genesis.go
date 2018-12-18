package types

import (
	"bytes"
	"encoding/binary"
	fmt "fmt"
	"math"
	"strings"

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
	// ConsensusMax is the max size of the Consensus field of ChainID.
	ConsensusMax = 10

	cidMarshal = iota
	cidUnmarshal
)

type errCidCodec struct {
	codec int
	field string
	err   error
}

func (e errCidCodec) Error() string {
	kind := "unmarshal"
	if e.codec == cidMarshal {
		kind = "marshal"
	}
	return fmt.Sprintf("failed to %s %s - %s", kind, e.field, e.err.Error())
}

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
func (cid *ChainID) Bytes() ([]byte, error) {
	var w bytes.Buffer

	// warning: when any field added to ChainID, the corresponding
	// serialization code must be written here.
	if err := binary.Write(&w, binary.LittleEndian, cid.Version); err != nil {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "version",
			err:   err,
		}
	}
	if err := binary.Write(&w, binary.LittleEndian, cid.PublicNet); err != nil {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "publicnet",
			err:   err,
		}
	}
	if err := binary.Write(&w, binary.LittleEndian, cid.MainNet); err != nil {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "mainnet",
			err:   err,
		}
	}

	if len(cid.Magic) > MagicMax || len(cid.Consensus) > ConsensusMax {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "magic/consensus",
			err: fmt.Errorf(
				"too large magic or consensus (size limit: magic <= %v, consensus <= %v)",
				MagicMax, ConsensusMax),
		}
	}

	magicConsensus := fmt.Sprintf("%s/%s", cid.Magic, cid.Consensus)
	if err := binary.Write(&w, binary.LittleEndian, []byte(magicConsensus)); err != nil {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "magic/consensus",
			err:   err,
		}
	}

	return w.Bytes(), nil
}

// Read deserialize data as a ChainID.
func (cid *ChainID) Read(data []byte) error {
	r := bytes.NewBuffer(data)

	// warning: when any field added to ChainID, the corresponding
	// deserialization code must be written here.
	if err := binary.Read(r, binary.LittleEndian, &cid.Version); err != nil {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "version",
			err:   err,
		}
	}
	if err := binary.Read(r, binary.LittleEndian, &cid.PublicNet); err != nil {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "publicnet",
			err:   err,
		}
	}
	if err := binary.Read(r, binary.LittleEndian, &cid.MainNet); err != nil {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "mainnet",
			err:   err,
		}
	}

	mc := strings.Split(string(r.Bytes()), "/")
	if len(mc) != 2 {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "magic/consensus",
			err:   fmt.Errorf("too many fields: %s", mc),
		}
	}
	cid.Magic, cid.Consensus = mc[0], mc[1]

	if len(cid.Magic) > MagicMax || len(cid.Consensus) > ConsensusMax {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "magic/consensus",
			err: fmt.Errorf(
				"too large magic or consensus (size limit: magic <= %v, consensus <= %v)",
				MagicMax, ConsensusMax),
		}
	}

	return nil
}

// AsDefault set *cid to the default chaind id (cid must be a valid pointer).
func (cid *ChainID) AsDefault() {
	*cid = defaultChainID
}

// Equals reports wheter cid equals rhs or not.
func (cid *ChainID) Equals(rhs *ChainID) bool {
	var (
		lVal, rVal []byte
		err        error
	)

	if lVal, err = cid.Bytes(); err != nil {
		return false
	}
	if rVal, err = rhs.Bytes(); err != nil {
		return false
	}

	return bytes.Compare(lVal, rVal) == 0
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
func (g *Genesis) ChainID() ([]byte, error) {
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

// ConsensusType retruns g.ID.ConsensusType.
func (g Genesis) ConsensusType() string {
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
