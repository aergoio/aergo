package jsonrpc

import (
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
)

func ConvBlock(msg *types.Block) *InOutBlock {
	if msg == nil {
		return nil
	}

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
	Hash   string           `json:"hash,omitempty"`
	Header InOutBlockHeader `json:"header,omitempty"`
	Body   InOutBlockBody   `json:"body,omitempty"`
}

func ConvBlockHeader(msg *types.BlockHeader) *InOutBlockHeader {
	if msg == nil {
		return nil
	}

	bh := &InOutBlockHeader{}
	bh.ChainID = base58.Encode(msg.GetChainID())
	bh.Version = types.DecodeChainIdVersion(msg.GetChainID())
	bh.PrevBlockHash = base58.Encode(msg.GetPrevBlockHash())
	bh.BlockNo = msg.GetBlockNo()
	bh.Timestamp = msg.GetTimestamp()
	bh.BlockRootHash = base58.Encode(msg.GetBlocksRootHash())
	bh.TxsRootHash = base58.Encode(msg.GetTxsRootHash())
	bh.ReceiptsRootHash = base58.Encode(msg.GetReceiptsRootHash())
	bh.Confirms = msg.GetConfirms()
	bh.PubKey = base58.Encode(msg.GetPubKey())
	if bpid, err := msg.BPID(); err == nil {
		bh.PeerID = bpid.String()
	}
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
	ChainID          string `json:"chainID,omitempty"`
	Version          int32  `json:"version,omitempty"`
	PrevBlockHash    string `json:"prevBlockHash,omitempty"`
	BlockNo          uint64 `json:"blockNo,omitempty"`
	Timestamp        int64  `json:"timestamp,omitempty"`
	BlockRootHash    string `json:"blockRootHash,omitempty"`
	TxsRootHash      string `json:"txsRootHash,omitempty"`
	ReceiptsRootHash string `json:"receiptsRootHash,omitempty"`
	Confirms         uint64 `json:"confirms,omitempty"`
	PubKey           string `json:"pubKey,omitempty"`
	PeerID           string `json:"peerID,omitempty"`
	CoinbaseAccount  string `json:"coinbaseAccount,omitempty"`
	Sign             string `json:"sign,omitempty"`
	Consensus        string `json:"consensus,omitempty"`
}

func ConvBlockBody(msg *types.BlockBody) *InOutBlockBody {
	if msg == nil {
		return nil
	}

	bb := &InOutBlockBody{}
	bb.Txs = make([]*InOutTx, len(msg.Txs))
	for i, tx := range msg.Txs {
		bb.Txs[i] = ConvTx(tx, Base58)
	}
	return bb
}

type InOutBlockBody struct {
	Txs []*InOutTx `json:"txs,omitempty"`
}

func ConvBlockIdx(msg *types.NewBlockNotice) *InOutBlockIdx {
	if msg == nil {
		return nil
	}

	return &InOutBlockIdx{
		BlockHash: base58.Encode(msg.GetBlockHash()),
		BlockNo:   msg.GetBlockNo(),
	}
}

type InOutBlockIdx struct {
	BlockHash string `json:"blockHash,omitempty"`
	BlockNo   uint64 `json:"blockNo,omitempty"`
}

func ConvBlockHeaderList(msg *types.BlockHeaderList) *InOutBlockHeaderList {
	if msg == nil {
		return nil
	}

	b := &InOutBlockHeaderList{}
	b.Blocks = make([]*InOutBlock, len(msg.Blocks))
	for i, block := range msg.Blocks {
		b.Blocks[i] = ConvBlock(block)
	}
	return b
}

type InOutBlockHeaderList struct {
	Blocks []*InOutBlock
}
