/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/rs/zerolog"
)

type LogTxHash struct {
	*Tx
}

func (t LogTxHash) MarshalZerologObject(e *zerolog.Event) {
	e.Str("txID", enc.ToString(t.Hash))
}

type LogTx struct {
	*Tx
}

func (t LogTx) MarshalZerologObject(e *zerolog.Event) {
	e.Str("txID", enc.ToString(t.GetHash())).Str("account", enc.ToString(t.Body.Account)).Uint64("nonce", t.Body.Nonce)
}

type LogTrsactions struct {
	TXs   []Transaction
	Limit int
}

func (l LogTrsactions) MarshalZerologArray(a *zerolog.Array) {
	size := len(l.TXs)
	if size > l.Limit {
		size = l.Limit - 1
		for _, tr := range l.TXs[:size] {
			marshalTrx(tr, a)
		}
		a.Str(fmt.Sprintf("(and %d more)", len(l.TXs)-size))
	} else {
		for _, tr := range l.TXs {
			marshalTrx(tr, a)
		}
	}
}

func marshalTrx(tr Transaction, a *zerolog.Array) {
	if tr == nil {
		a.Str("(nil)")
	} else {
		if tx := tr.GetTx(); tx == nil {
			a.Str("(nil tx)")
		} else {
			a.Object(LogTx{tx})
		}
	}
}

type LogBase58 []byte

func (t LogBase58) String() string {
	return enc.ToString(t)
}

func (t LogBase58) MarshalZerologObject(e *zerolog.Event) {
	e.Str("b58", enc.ToString(t))
}
