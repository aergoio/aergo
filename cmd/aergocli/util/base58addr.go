package util

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/aergoio/aergo/types"
	"github.com/anaskhan96/base58check"
	"github.com/mr-tron/base58/base58"
)

type InOutTx struct {
	Hash string
	Body *InOutTxBody
}

type InOutTxBody struct {
	Nonce     uint64
	Account   string
	Recipient string
	Amount    string
	Payload   string
	Limit     uint64
	Price     string
	Type      types.TxType
	Sign      string
}

type InOutTxIdx struct {
	BlockHash string
	Idx       int32
}

type InOutTxInBlock struct {
	TxIdx *InOutTxIdx
	Tx    *InOutTx
}

type InOutBlockHeader struct {
	ChainID          string
	PrevBlockHash    string
	BlockNo          uint64
	Timestamp        int64
	BlockRootHash    string
	TxRootHash       string
	ReceiptsRootHash string
	Confirms         uint64
	PubKey           string
	Sign             string
	CoinbaseAccount  string
}

type InOutBlockBody struct {
	Txs []*InOutTx
}

type InOutBlock struct {
	Hash   string
	Header InOutBlockHeader
	Body   InOutBlockBody
}

type InOutBlockIdx struct {
	BlockHash string
	BlockNo   uint64
}

type InOutPeerAddress struct {
	Address string
	Port    string
	PeerId  string
}

type InOutPeer struct {
	Address   InOutPeerAddress
	BestBlock InOutBlockIdx
	LastCheck time.Time
	State     string
	Hidden    bool
	Self      bool
}

func FillTxBody(source *InOutTxBody, target *types.TxBody) error {
	var err error
	if source == nil {
		return errors.New("tx body is empty")
	}
	target.Nonce = source.Nonce
	if source.Account != "" {
		target.Account, err = types.DecodeAddress(source.Account)
		if err != nil {
			return err
		}
	}
	if source.Recipient != "" {
		target.Recipient, err = types.DecodeAddress(source.Recipient)
		if err != nil {
			return err
		}
	}
	if source.Amount != "" {
		amount, err := ParseUnit(source.Amount)
		if err != nil {
			return err
		}
		target.Amount = amount.Bytes()
	}
	if source.Payload != "" {
		target.Payload, err = base58.Decode(source.Payload)
		if err != nil {
			return err
		}
	}
	target.Limit = source.Limit
	if source.Price != "" {
		price, err := ParseUnit(source.Price)
		if err != nil {
			return err
		}
		target.Price = price.Bytes()
	}
	if source.Sign != "" {
		target.Sign, err = base58.Decode(source.Sign)
		if err != nil {
			return err
		}
	}
	target.Type = source.Type
	return nil
}

func ParseBase58Tx(jsonTx []byte) ([]*types.Tx, error) {
	var inputlist []InOutTx
	err := json.Unmarshal([]byte(jsonTx), &inputlist)
	if err != nil {
		var input InOutTx
		err = json.Unmarshal([]byte(jsonTx), &input)
		if err != nil {
			return nil, err
		}
		inputlist = append(inputlist, input)
	}
	txs := make([]*types.Tx, len(inputlist))
	for i, in := range inputlist {
		tx := &types.Tx{Body: &types.TxBody{}}
		if in.Hash != "" {
			tx.Hash, err = base58.Decode(in.Hash)
			if err != nil {
				return nil, err
			}
		}
		err = FillTxBody(in.Body, tx.Body)
		if err != nil {
			return nil, err
		}
		txs[i] = tx
	}

	return txs, nil
}

func ParseBase58TxBody(jsonTx []byte) (*types.TxBody, error) {
	body := &types.TxBody{}
	in := &InOutTxBody{}

	err := json.Unmarshal([]byte(jsonTx), in)
	if err != nil {
		return nil, err
	}

	err = FillTxBody(in, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ConvTx(tx *types.Tx) *InOutTx {
	out := &InOutTx{Body: &InOutTxBody{}}
	if tx == nil {
		return out
	}
	out.Hash = base58.Encode(tx.Hash)
	out.Body.Nonce = tx.Body.Nonce
	if tx.Body.Account != nil {
		out.Body.Account = types.EncodeAddress(tx.Body.Account)
	}
	if tx.Body.Recipient != nil {
		out.Body.Recipient = types.EncodeAddress(tx.Body.Recipient)
	}
	out.Body.Amount = new(big.Int).SetBytes(tx.Body.Amount).String()
	out.Body.Payload = base58.Encode(tx.Body.Payload)
	out.Body.Limit = tx.Body.Limit
	out.Body.Price = new(big.Int).SetBytes(tx.Body.Price).String()
	out.Body.Sign = base58.Encode(tx.Body.Sign)
	out.Body.Type = tx.Body.Type
	return out
}

func ConvTxInBlock(txInBlock *types.TxInBlock) *InOutTxInBlock {
	out := &InOutTxInBlock{TxIdx: &InOutTxIdx{}, Tx: &InOutTx{}}
	out.TxIdx.BlockHash = base58.Encode(txInBlock.GetTxIdx().GetBlockHash())
	out.TxIdx.Idx = txInBlock.GetTxIdx().GetIdx()
	out.Tx = ConvTx(txInBlock.GetTx())
	return out
}

func ConvBlock(b *types.Block) *InOutBlock {
	out := &InOutBlock{}
	if b != nil {
		out.Hash = base58.Encode(b.Hash)
		out.Header.ChainID = base58.Encode(b.GetHeader().GetChainID())
		out.Header.PrevBlockHash = base58.Encode(b.GetHeader().GetPrevBlockHash())
		out.Header.BlockNo = b.GetHeader().GetBlockNo()
		out.Header.Timestamp = b.GetHeader().GetTimestamp()
		out.Header.BlockRootHash = base58.Encode(b.GetHeader().GetBlocksRootHash())
		out.Header.TxRootHash = base58.Encode(b.GetHeader().GetTxsRootHash())
		out.Header.ReceiptsRootHash = base58.Encode(b.GetHeader().GetReceiptsRootHash())
		out.Header.Confirms = b.GetHeader().GetConfirms()
		out.Header.PubKey = base58.Encode(b.GetHeader().GetPubKey())
		out.Header.Sign = base58.Encode(b.GetHeader().GetSign())
		out.Header.CoinbaseAccount = base58.Encode(b.GetHeader().GetCoinbaseAccount())
		if b.Body != nil {
			for _, tx := range b.Body.Txs {
				out.Body.Txs = append(out.Body.Txs, ConvTx(tx))
			}
		}
	}
	return out
}

func ConvPeer(p *types.Peer) *InOutPeer {
	out := &InOutPeer{}
	out.Address.Address = p.GetAddress().GetAddress()
	out.Address.Port = strconv.Itoa(int(p.GetAddress().GetPort()))
	out.Address.PeerId = base58.Encode(p.GetAddress().GetPeerID())
	out.LastCheck = time.Unix(0, p.GetLashCheck())
	out.BestBlock.BlockNo = p.GetBestblock().GetBlockNo()
	out.BestBlock.BlockHash = base58.Encode(p.GetBestblock().GetBlockHash())
	out.State = types.PeerState(p.State).String()
	out.Hidden = p.Hidden
	out.Self = p.Selfpeer
	return out
}

func ConvBlockchainStatus(in *types.BlockchainStatus) string {
	out := &InOutBlockchainStatus{}
	if in == nil {
		return ""
	}
	out.Hash = base58.Encode(in.BestBlockHash)
	out.Height = in.BestHeight
	jsonout, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(jsonout)
}

func TxConvBase58Addr(tx *types.Tx) string {
	return toString(ConvTx(tx))
}

func TxInBlockConvBase58Addr(txInBlock *types.TxInBlock) string {
	return toString(ConvTxInBlock(txInBlock))
}

func BlockConvBase58Addr(b *types.Block) string {
	return toString(ConvBlock(b))
}

func PeerListToString(p *types.PeerList) string {
	peers := []*InOutPeer{}
	for _, peer := range p.GetPeers() {
		peers = append(peers, ConvPeer(peer))
	}
	return toString(peers)
}

func toString(out interface{}) string {
	jsonout, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return ""
	}
	return string(jsonout)
}

const CodeVersion = 0xC0

func EncodeCode(code []byte) string {
	encoded, _ := base58check.Encode(fmt.Sprintf("%x", CodeVersion), hex.EncodeToString(code))
	return encoded
}

func DecodeCode(encodedCode string) ([]byte, error) {
	decodedString, err := base58check.Decode(encodedCode)
	if err != nil {
		return nil, err
	}
	decodedBytes, err := hex.DecodeString(decodedString)
	if err != nil {
		return nil, err
	}
	version := decodedBytes[0]
	if version != CodeVersion {
		return nil, errors.New("Invalid code version")
	}
	decoded := decodedBytes[1:]
	return decoded, nil
}
