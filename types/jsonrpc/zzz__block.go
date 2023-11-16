package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvBlock(msg *types.Block) *InOutBlock {
	b := &InOutBlock{}
	b.Hash = base58.Encode(msg.Hash)
	if msg.Header != nil {
		b.Header = *ConvBlockHeader(msg.Header)
	}
	if msg.Body != nil {
		b.Body = *ConvBlockBody(msg.Body)
	}
	return b
}

type InOutBlock struct {
	Hash   string
	Header InOutBlockHeader
	Body   InOutBlockBody
}

func ConvBlockHeader(msg *types.BlockHeader) *InOutBlockHeader {
	bh := &InOutBlockHeader{}
	bh.ChainID = base58.Encode(msg.GetChainID())
	bh.Version = types.DecodeChainIdVersion(msg.GetChainID())
	bh.PrevBlockHash = base58.Encode(msg.GetPrevBlockHash())
	bh.BlockNo = msg.GetBlockNo()
	bh.Timestamp = msg.GetTimestamp()
	bh.BlockRootHash = base58.Encode(msg.GetBlocksRootHash())
	bh.TxRootHash = base58.Encode(msg.GetTxsRootHash())
	bh.ReceiptsRootHash = base58.Encode(msg.GetReceiptsRootHash())
	bh.Confirms = msg.GetConfirms()
	bh.PubKey = base58.Encode(msg.GetPubKey())
	bh.Sign = base58.Encode(msg.GetSign())
	if msg.GetCoinbaseAccount() != nil {
		bh.CoinbaseAccount = types.EncodeAddress(msg.GetCoinbaseAccount())
	}
	if consensus := msg.GetConsensus(); consensus != nil {
		bh.Consensus = types.EncodeAddress(consensus)
	}
	return bh
}

type InOutBlockHeader struct {
	ChainID          string
	Version          int32
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
	Consensus        string
}

func ConvBlockBody(msg *types.BlockBody) *InOutBlockBody {
	bb := &InOutBlockBody{}
	bb.Txs = make([]*InOutTx, len(msg.Txs))
	for i, tx := range msg.Txs {
		bb.Txs[i] = ConvTx(tx, Base58)
	}
	return bb
}

type InOutBlockBody struct {
	Txs []*InOutTx
}

func ConvBlockIdx(msg *types.NewBlockNotice) *InOutBlockIdx {
	bh := &InOutBlockIdx{}
	bh.BlockNo = msg.GetBlockNo()
	bh.BlockHash = base58.Encode(msg.GetBlockHash())
	return bh
}

type InOutBlockIdx struct {
	BlockHash string
	BlockNo   uint64
}
