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

type LogStringersMarshaller struct {
	arr   []fmt.Stringer
	limit int
}

func NewLogStringersMarshaller(arr []fmt.Stringer, limit int) *LogStringersMarshaller {
	return &LogStringersMarshaller{arr: arr, limit: limit}
}

func (m *LogStringersMarshaller) MarshalZerologArray(a *zerolog.Array) {
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

type LogPeerMetasMarshaller struct {
	metas []p2pcommon.PeerMeta
	limit int
}

func NewLogPeerMetasMarshaller(metas []p2pcommon.PeerMeta, limit int) *LogPeerMetasMarshaller {
	return &LogPeerMetasMarshaller{metas: metas, limit: limit}
}

func (m *LogPeerMetasMarshaller) MarshalZerologArray(a *zerolog.Array) {
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

// LogB58EncMarshaller is zerolog array marshaller which print bytes array to base58 encoded string.
type LogB58EncMarshaller struct {
	arr   [][]byte
	limit int
}

func NewLogB58EncMarshaller(arr [][]byte, limit int) *LogB58EncMarshaller {
	return &LogB58EncMarshaller{arr: arr, limit: limit}
}

func (m *LogB58EncMarshaller) MarshalZerologArray(a *zerolog.Array) {
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
