package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	fmt "fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/aergoio/aergo/internal/common"
)

const (
	blockVersionNil = math.MinInt32
	devChainMagic   = "dev.chain"
)

var (
	nilChainID = ChainID{
		Version:     0,
		Magic:       "",
		PublicNet:   false,
		MainNet:     false,
		Consensus:   "",
		CoinbaseFee: DefaultCoinbaseFee,
	}

	defaultChainID = ChainID{
		Version:     0,
		Magic:       devChainMagic,
		PublicNet:   false,
		MainNet:     false,
		Consensus:   "sbp",
		CoinbaseFee: DefaultCoinbaseFee,
	}
)

const (
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
	Version     int32  `json:"-"`
	PublicNet   bool   `json:"public"`
	MainNet     bool   `json:"mainnet"`
	CoinbaseFee string `json:"coinbasefee"`
	Magic       string `json:"magic"`
	Consensus   string `json:"consensus"`
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
	if n, ok := cid.GetCoinbaseFee(); !ok || n.Sign() < 0 {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "coinbasefee",
			err: fmt.Errorf(
				"coinbasefee is not proper number(%s)", cid.CoinbaseFee),
		}
	}
	if cid.PublicNet && cid.CoinbaseFee != DefaultCoinbaseFee {
		return nil, errCidCodec{
			codec: cidMarshal,
			field: "coinbasefee",
			err: fmt.Errorf(
				"coinbasefee for mainnet should be %s not %s", DefaultCoinbaseFee, cid.CoinbaseFee),
		}
	}

	others := fmt.Sprintf("%s/%s/%s", cid.CoinbaseFee, cid.Magic, cid.Consensus)
	if err := binary.Write(&w, binary.LittleEndian, []byte(others)); err != nil {
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
	if len(mc) != 3 {
		return errCidCodec{
			codec: cidUnmarshal,
			field: "coinbasefee/magic/consensus",
			err:   fmt.Errorf("wrong number of fields: %s", mc),
		}
	}
	cid.CoinbaseFee, cid.Magic, cid.Consensus = mc[0], mc[1], mc[2]

	if _, ok := cid.GetCoinbaseFee(); !ok {
		return errCidCodec{
			codec: cidMarshal,
			field: "coinbasefee",
			err: fmt.Errorf(
				"coinbasefee is not proper number(%s)", cid.CoinbaseFee),
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

func (cid *ChainID) GetCoinbaseFee() (*big.Int, bool) {
	return new(big.Int).SetString(cid.CoinbaseFee, 10)
}

// ToJSON returns a JSON encoded string of cid.
func (cid ChainID) ToJSON() string {
	if b, err := json.Marshal(cid); err == nil {
		return string(b)
	}
	return ""
}

// Genesis represents genesis block
type Genesis struct {
	ID        ChainID           `json:"chain_id,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
	Balance   map[string]string `json:"balance"`
	BPs       []string          `json:"bps"`

	// followings are for internal use only
	totalBalance *big.Int
	block        *Block
}

// Block returns Block corresponding to g. If g.block == nil, it genreates a
// genesis block before it returns.
func (g *Genesis) Validate() error {
	_, err := g.ChainID()
	if err != nil {
		return err
	}
	//TODO check BP count
	return nil
}

// Block returns Block corresponding to g.
func (g *Genesis) Block() *Block {
	if g.block == nil {
		g.SetBlock(NewBlock(nil, nil, nil, nil, nil, g.Timestamp))
		if id, err := g.ID.Bytes(); err == nil {
			g.block.SetChainID(id)
		}
	}
	return g.block
}

// AddBalance adds bal to g.totalBalance.
func (g *Genesis) AddBalance(bal *big.Int) {
	if g.totalBalance == nil {
		g.totalBalance = big.NewInt(0)
	}
	g.totalBalance.Add(g.totalBalance, bal)
}

// TotalBalance returns the total initial balance of the chain.
func (g *Genesis) TotalBalance() *big.Int {
	return g.totalBalance
}

// SetTotalBalance sets g.totalBalance to v if g.totalBlance has no valid
// value (nil).
func (g *Genesis) SetTotalBalance(v []byte) {
	if g.totalBalance == nil {
		g.totalBalance = big.NewInt(0).SetBytes(v)
	}
}

// SetBlock sets g.block to block if g.block == nil.
func (g *Genesis) SetBlock(block *Block) {
	if g.block == nil {
		g.block = block
	}
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

// PublicNet reports whether g corresponds to PublicNet.
func (g Genesis) PublicNet() bool {
	return g.ID.PublicNet
}

func (g Genesis) IsAergoPublicChain() bool {

	testNetCid := GetTestNetGenesis().ID
	if testNetCid.Equals(&g.ID) {
		return true
	}
	//TODO check MainNetChainID later
	return false
}

func (g Genesis) HasDevChainID() bool {
	if g.ID.Magic == devChainMagic {
		return true
	}
	return false
}

func (g Genesis) HasPrivateChainID() bool {
	if g.IsAergoPublicChain() || g.HasDevChainID() {
		return false
	}
	return true
}

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *Genesis {
	return &Genesis{
		ID:        defaultChainID,
		Timestamp: time.Now().UnixNano(),
		block:     nil,
	} //TODO embed MAINNET genesis block
}

func GetTestNetGenesis() *Genesis {
	if bs, err := hex.DecodeString(TestNetGenesis); err == nil {
		var g Genesis
		if err := json.Unmarshal(bs, &g); err == nil {
			return &g
		}
	}
	return nil
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
