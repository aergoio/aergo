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

	"github.com/aergoio/aergo/internal/common"
)

const (
	// DefaultSeed is temporary const to create same genesis block with no
	// configuration. This is a UNIX timestamp (2018-07-06 10:00:00 +0900 KST)
	// in nanoseconds.
	DefaultSeed     = 1530838800000000000
	blockVersionNil = math.MinInt32
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
		Magic:       "",
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

// GetDefaultGenesis returns default genesis structure
func GetDefaultGenesis() *Genesis {
	return &Genesis{
		ID:        defaultChainID,
		Timestamp: DefaultSeed,
		block:     nil,
	} //TODO embed MAINNET genesis block
}

func GetTestNetGenesis() *Genesis {
	if bs, err := hex.DecodeString(PredefinedTestNet); err == nil {
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

var PredefinedTestNet = "7b22636861696e5f6964223a7b227075626c6963223a747275652c226d61696e6e6574223a66616c73652c22636f696e62617365666565223a2231303030303030303030222c226d61676963223a22746573746e65742e616572676f2e696f222c22636f6e73656e737573223a2264706f73227d2c2274696d657374616d70223a313534353938303430303030303030303030302c2262616c616e6365223a7b22416d4c715a466e774d4c714c6735664d7368677a6d667677425038756959476766563374425a416d33365476376a465963733466223a22313030303030303030303030303030303030303030303030222c22416d4c735366786f396151525a4a42764d426f4c466239515a4142514b32526947335571314a4268794166624459506633314a32223a22313030303030303030303030303030303030303030303030222c22416d4c7654415853545641733774674153576e6d5a776774797137634c6474616f4e7178726d31314b39314a7063335872486578223a22313030303030303030303030303030303030303030303030222c22416d4d4b334c5a6952316f45663636787a586972376d4135535556564853696e5755596d6835467775656f566d6369483343754a223a22343938373030303030303030303030303030303030303030303030222c22416d4d4c464e733166356b4678786f4732507138776b4669356d4a4154517665375a5050425a6e6748526a6f326d795534316b36223a22313030303030303030303030303030303030303030303030222c22416d4d5231434c484e6e4243713769745935585258344c714c73355665316e5a4b7841634252565a74776373526472797446716b223a22313030303030303030303030303030303030303030303030222c22416d4d766d753648785268594d6a5667696834735853597944694d38487075765a466e50395648557455643935715455454e5236223a22313030303030303030303030303030303030303030303030222c22416d4e4566734b74317632676f73627536654d673766474263686b4a7a7645636d624738413445654d7975734155334d33386d6b223a22313030303030303030303030303030303030303030303030222c22416d4e4a5268346e745157543134356a5177444851477548596441546d4d57434562454e5732444e376d7a746672357071506250223a22313030303030303030303030303030303030303030303030222c22416d4e536d6b4474777753537247486b34455038794c5037596143363239666a6e50673937374a4a664c705465397a5874707364223a22313030303030303030303030303030303030303030303030222c22416d4e6d6b4b584b677758474c4263356f466869314d75784e786e4a4c7776575550485635777235756e48564e7758314d665865223a22313030303030303030303030303030303030303030303030222c22416d50357659586548426a47757671446147725a426f48795448454a63557a7455447654616d32487a717641396a736571656b64223a22313030303030303030303030303030303030303030303030222c22416d504a524c48444b747a4c707361433875626d5075526b786e4d437942537135774277594e444436444a646769526841685952223a22313030303030303030303030303030303030303030303030222c22416d5078445a596d633766366354457a763872546b62447859747a69546b525145426243573972354767566e7470536d67585762223a22313030303030303030303030303030303030303030303030227d2c22627073223a5b2231365569753248416d416f6b5941744c625a784a4150526770326a43633462443335634a4439323174727155414e6835395263346e222c2231365569753248416d347859744773716b3757474b5578723870726656704a323568443233415133426536616e454c394b786b6777222c2231365569753248416d47694a325167564157484d55747a4c4b4b4e4d356546554a33447333464e376e594a71316d484e355a506a39222c2231365569753248416d524e59364345444d514b75556b5562697837514470703166464167337952564d6f4a4a444e7a7462477a6448222c2231365569753248416d31384c795462395757564e57636a35476e383366555034636b7178677939417a48677a59735870695a384a41222c2231365569753248416d4e6165727436487353615731626273384a6233685571326e57366979694d676a31684277546735466b427350222c2231365569753248416d5636695647754e3331735a547a32474469634650704272366561486e316d564d343939424777534266364e62222c2231365569753248416d376d763831676b654c4d747732376235646f58487668454c516e487766486842665a4c4771724c6a5a4c565a222c2231365569753248416d344e747663714e4b465755317a336f4d5242335a756765386a765a693379626e367842745a6b4c4145765557222c2231365569753248416d50574b776142456f4463744e513969766966723776365a636f3645416a6676746353317764385472344c5048222c2231365569753248416d526167343263616b7966696b6555544367544c6f47446e50755a646b695046775a733575367a7034444a784d222c2231365569753248416d38474b63766b57396e48474d596638767a466e4564516e6b55357666736b4e6e4c784b3155754a6e32705334222c2231365569753248416d375446573150617658474e4b447779654d614c3943785956537059707a73387432425656336e663871744278225d7d"
