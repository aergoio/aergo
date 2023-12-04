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
	Blocks []*InOutBlock `json:"blocks"`
}

func ConvBlockBodyPaged(msg *types.BlockBodyPaged) *InOutBlockBodyPaged {
	if msg == nil {
		return nil
	}

	bbp := &InOutBlockBodyPaged{}
	bbp.Body = ConvBlockBody(msg.GetBody())
	bbp.Total = msg.GetTotal()
	bbp.Offset = msg.GetOffset()
	bbp.Size = msg.GetSize()

	return bbp
}

type InOutBlockBodyPaged struct {
	Total  uint32          `json:"total,omitempty"`
	Offset uint32          `json:"offset,omitempty"`
	Size   uint32          `json:"size,omitempty"`
	Body   *InOutBlockBody `json:"body,omitempty"`
}

func ConvBlockMetadata(msg *types.BlockMetadata) *InOutBlockMetadata {
	if msg == nil {
		return nil
	}

	bbm := &InOutBlockMetadata{}
	bbm.Hash = base58.Encode(msg.Hash)
	bbm.Header = ConvBlockHeader(msg.GetHeader())
	bbm.Txcount = msg.GetTxcount()
	bbm.Size = msg.GetSize()

	return bbm
}

type InOutBlockMetadata struct {
	Hash    string            `json:"hash,omitempty"`
	Header  *InOutBlockHeader `json:"header,omitempty"`
	Txcount int32             `json:"txcount,omitempty"`
	Size    int64             `json:"size,omitempty"`
}

func ConvListBlockMetadata(msg *types.BlockMetadataList) *InOutBlockMetadataList {
	if msg == nil {
		return nil
	}

	bbml := &InOutBlockMetadataList{}
	bbml.Blocks = make([]*InOutBlockMetadata, len(msg.Blocks))
	for i, block := range msg.Blocks {
		bbml.Blocks[i] = ConvBlockMetadata(block)
	}

	return bbml
}

type InOutBlockMetadataList struct {
	Blocks []*InOutBlockMetadata `json:"blocks,omitempty"`
}

func ConvBlockTransactionCount(msg *types.Block) *InOutBlockTransactionCount {
	if msg == nil {
		return nil
	}

	btc := &InOutBlockTransactionCount{}
	if msg.Body.Txs != nil {
		btc.Count = len(msg.Body.Txs)
	} else {
		btc.Count = 0
	}

	return btc
}

type InOutBlockTransactionCount struct {
	Count int `json:"count,omitempty"`
}
