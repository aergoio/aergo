/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/rs/zerolog"
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
	arr []*Block
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

func (m *GetTransactionsRequest) MarshalZerologObject(e *zerolog.Event) {
	e.Array("hashes", NewLogB58EncMarshaller(m.Hashes, 10))
}


func (m *GetBlockResponse) MarshalZerologObject(e *zerolog.Event) {
	e.Str("status", m.Status.String()).Bool("hasNext", m.HasNext).Array("hashes", LogBlockHashMarshaller{m.Blocks, 10})
}


func (m *NewTransactionsNotice) MarshalZerologObject(e *zerolog.Event) {
	e.Array("hashes", NewLogB58EncMarshaller(m.TxHashes, 10))
}

func (m *BlockProducedNotice)  MarshalZerologObject(e *zerolog.Event) {
	e.Str("bp", enc.ToString(m.ProducerID)).Uint64("blk_no", m.BlockNo).Str("blk_hash", enc.ToString(m.Block.Hash))
}


func (m *Ping)  MarshalZerologObject(e *zerolog.Event) {
	e.Str("blk_hash", enc.ToString(m.BestBlockHash)).Uint64("blk_no", m.BestHeight)
}
