// blockchain

// wording directory

// height?

// oo

package context

import (
	"strconv"

	"github.com/aergoio/aergo/contract"
)

var CurrentCtx *context

var GitHash = "?"

func init() {
	Reset()
}

type context struct {
	chain *contract.DummyChain
}

func Reset() {
	chain, err := contract.LoadDummyChain(contract.OnPubNet)
	if err != nil {
		panic(err)
	}
	CurrentCtx = &context{
		chain: chain,
	}
}

func LivePrefix() (string, bool) {
	height := strconv.FormatUint(CurrentCtx.chain.BestBlockNo(), 10)
	ret := height + "> "
	return ret, true
}

func Get() *contract.DummyChain {
	return CurrentCtx.chain
}
