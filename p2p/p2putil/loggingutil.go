/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/rs/zerolog"
)

type LogStringersMarshaler struct {
	arr   []fmt.Stringer
	limit int
}

func NewLogStringersMarshaler(arr []fmt.Stringer, limit int) *LogStringersMarshaler {
	return &LogStringersMarshaler{arr: arr, limit: limit}
}

func (m *LogStringersMarshaler) MarshalZerologArray(a *zerolog.Array) {
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

type LogPeerMetasMarshaler struct {
	metas []p2pcommon.PeerMeta
	limit int
}

func NewLogPeerMetasMarshaler(metas []p2pcommon.PeerMeta, limit int) *LogPeerMetasMarshaler {
	return &LogPeerMetasMarshaler{metas: metas, limit: limit}
}

func (m *LogPeerMetasMarshaler) MarshalZerologArray(a *zerolog.Array) {
	size := len(m.metas)
	if size > m.limit {
		for i := 0; i < m.limit-1; i++ {
			a.Str(ShortMetaForm(m.metas[i]))
		}
		a.Str(fmt.Sprintf("(and %d more)", size-m.limit+1))
	} else {
		for _, meta := range m.metas {
			a.Str(ShortMetaForm(meta))
		}
	}
}

// LogB58EncMarshaler is zerolog array marshaler which print bytes array to baase58 encoded string.
type LogB58EncMarshaler struct {
	arr   [][]byte
	limit int
}

func NewLogB58EncMarshaler(arr [][]byte, limit int) *LogB58EncMarshaler {
	return &LogB58EncMarshaler{arr: arr, limit: limit}
}

func (m *LogB58EncMarshaler) MarshalZerologArray(a *zerolog.Array) {
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
