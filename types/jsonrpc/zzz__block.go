package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

type InOutBlock struct {
	Hash   string
	Header InOutBlockHeader
	Body   InOutBlockBody
}

func (b *InOutBlock) FromProto(msg *types.Block) {
	b.Hash = base58.Encode(msg.Hash)
	if msg.Header != nil {
		b.Header.FromProto(msg.Header)
	}
	if msg.Body != nil {
		b.Body.FromProto(msg.Body)
	}
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

func (bh *InOutBlockHeader) FromProto(msg *types.BlockHeader) {
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
}

type InOutBlockBody struct {
	Txs []*InOutTx
}

func (bb *InOutBlockBody) FromProto(msg *types.BlockBody) {
	bb.Txs = make([]*InOutTx, len(msg.Txs))
	for i, tx := range msg.Txs {
		bb.Txs[i] = &InOutTx{}
		bb.Txs[i].FromProto(tx, Base58)
	}
}

type InOutBlockIdx struct {
	BlockHash string
	BlockNo   uint64
}

func (bh *InOutBlockIdx) FromProto(msg *types.NewBlockNotice) {
	bh.BlockNo = msg.GetBlockNo()
	bh.BlockHash = base58.Encode(msg.GetBlockHash())
}
