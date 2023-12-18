/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/rs/zerolog"
)

type LogTxHash struct {
	*Tx
}

func (t LogTxHash) MarshalZerologObject(e *zerolog.Event) {
	e.Str("txID", base58.Encode(t.Hash))
}

type LogTx struct {
	*Tx
}

func (t LogTx) MarshalZerologObject(e *zerolog.Event) {
	e.Str("txID", base58.Encode(t.GetHash())).Str("account", base58.Encode(t.Body.Account)).Uint64("nonce", t.Body.Nonce)
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

// LogBase58 is thin wrapper which show base58 encoded string of byte array
type LogBase58 []byte

func (t LogBase58) String() string {
	return base58.Encode(t)
}

// LogAddr is thin wrapper which show base58 encoded form of wallet or smart contract
type LogAddr Address

func (t LogAddr) String() string {
	return EncodeAddress(t)
}

type LogPeerShort PeerID

func (t LogPeerShort) String() string {
	// basically this function is same as function p2putils.ShortForm()
	pretty := PeerID(t).String()
	if len(pretty) > 10 {
		return fmt.Sprintf("%s*%s", pretty[:2], pretty[len(pretty)-6:])
	} else {
		return pretty
	}
}
