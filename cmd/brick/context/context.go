// blockchain

// wording directory

// height?

// oo

package context

import (
	"log"
	"strconv"

	"github.com/aergoio/aergo/contract"
)

var (
	currentCtx *context
	GitHash    = "?"
	privateNet bool
)

type context struct {
	chain *contract.DummyChain
}

func Open(private bool) {
	privateNet = private
	var (
		chain *contract.DummyChain
		err   error
	)
	if privateNet {
		chain, err = contract.LoadDummyChain()
	} else {
		chain, err = contract.LoadDummyChain(contract.OnPubNet)
	}
	if err != nil {
		panic(err)
	}
	currentCtx = &context{
		chain: chain,
	}
}

func Close() {
	if currentCtx != nil {
		currentCtx.chain.Release()
	}
}

func Reset() {
	Open(privateNet)
}

func LivePrefix() (string, bool) {
	height := strconv.FormatUint(currentCtx.chain.BestBlockNo(), 10)
	ret := height + "> "
	return ret, true
}

func Get() *contract.DummyChain {
	if currentCtx == nil {
		log.Fatal("fail to open chain")
	}
	return currentCtx.chain
}
