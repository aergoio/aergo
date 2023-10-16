/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/rs/zerolog"
)

const (
	LogChainID    = "chain_id"
	LogBlkHash    = "blk_hash"
	LogBlkNo      = "blk_no"
	LogRespStatus = "status"
	LogHasNext    = "has_next"
)

// LogB58EncMarshaller is zerolog array marshaller which print bytes array to base58 encoded string.
type LogB58EncMarshaller struct {
	arr   [][]byte
	limit int
}

func NewLogB58EncMarshaller(arr [][]byte, limit int) *LogB58EncMarshaller {
	return &LogB58EncMarshaller{arr: arr, limit: limit}
}

func (m LogB58EncMarshaller) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(enc.ToString(m.arr[i]))
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(enc.ToString(element))
		}
	}
}

type LogBlockHashMarshaller struct {
	arr   []*Block
	limit int
}

func (m LogBlockHashMarshaller) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(enc.ToString(m.arr[i].GetHash()))
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(enc.ToString(element.GetHash()))
		}
	}
}

type LogTxIDsMarshaller struct {
	arr   []TxID
	limit int
}

func NewLogTxIDsMarshaller(arr []TxID, limit int) *LogTxIDsMarshaller {
	return &LogTxIDsMarshaller{arr: arr, limit: limit}
}

func (m LogTxIDsMarshaller) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(m.arr[i].String())
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(element.String())
		}
	}
}

type RaftMbrsMarshaller struct {
	arr   []*MemberAttr
	limit int
}

func (m RaftMbrsMarshaller) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.arr)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(m.arr[i].GetName())
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, element := range m.arr {
			a.Str(element.GetName())
		}
	}
}

func (m *GoAwayNotice) MarshalZerologObject(e *zerolog.Event) {
	e.Str("reason", m.Message)
}

func (m *AddressesResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Int("count", len(m.Peers))
}

func (m *GetTransactionsRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Int("count", len(m.Hashes)).Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetTransactionsResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Bool(LogHasNext, m.HasNext).Int("count", len(m.Hashes)).Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetBlockRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetBlockResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Bool(LogHasNext, m.HasNext).Int("count", len(m.Blocks)).Array("hashes", LogBlockHashMarshaller{m.Blocks, 10})
}

func (m *NewTransactionsNotice) MarshalZerologObject(e *zerolog.Event) {
	e.Int("count", len(m.TxHashes)).Array("hashes", NewLogB58EncMarshaller(m.TxHashes, 10))
}

func (m *BlockProducedNotice) MarshalZerologObject(e *zerolog.Event) {
	e.Str("bp", enc.ToString(m.ProducerID)).Uint64(LogBlkNo, m.BlockNo).Str(LogBlkHash, enc.ToString(m.Block.Hash))
}

func (m *Ping) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogBlkHash, enc.ToString(m.BestBlockHash)).Uint64(LogBlkNo, m.BestHeight)
}

func (m *GetHashesRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Str("prev_hash", enc.ToString(m.PrevHash)).Uint64("prev_no", m.PrevNumber)
}

func (m *GetHashesResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Bool(LogHasNext, m.HasNext).Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetBlockHeadersRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogBlkHash, enc.ToString(m.Hash)).Uint64(LogBlkNo, m.Height).Bool("ascending", m.Asc).Uint32("size", m.Size)
}

func (m *GetBlockHeadersResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Bool(LogHasNext, m.HasNext).Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetHashByNo) MarshalZerologObject(e *zerolog.Event) {
	e.Uint64(LogBlkNo, m.BlockNo)
}

func (m *GetHashByNoResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Str(LogBlkHash, enc.ToString(m.BlockHash))
}

func (m *GetAncestorRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}

func (m *GetAncestorResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogRespStatus, m.Status.String()).Str(LogBlkHash, enc.ToString(m.AncestorHash)).Uint64(LogBlkNo, m.AncestorNo)
}

func (m *GetClusterInfoRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Str("best_hash", enc.ToString(m.BestBlockHash))
}

func (m *GetClusterInfoResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str(LogChainID, enc.ToString(m.ChainID)).Str("err", m.Error).Array("members", RaftMbrsMarshaller{arr: m.MbrAttrs, limit: 10}).Uint64("cluster_id", m.ClusterID)
}

func (m *IssueCertificateResponse) MarshalZerologObject(e *zerolog.Event) {
	if m.Certificate != nil {
		e.Str("cert", m.Certificate.String())
	}
}

func (m *CertificateRenewedNotice) MarshalZerologObject(e *zerolog.Event) {
	if m.Certificate != nil {
		e.Str("cert", m.Certificate.String())
	}
}
