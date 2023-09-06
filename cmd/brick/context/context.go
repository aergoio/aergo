// blockchain

// wording directory

// height?

// oo

package context

import (
	"log"
	"strconv"

	"github.com/aergoio/aergo/v2/contract/vm_dummy"
)

var (
	currentCtx *context
	GitHash    = "?"
	privateNet bool
)

type context struct {
	chain *vm_dummy.DummyChain
}

func Open(private bool) {
	privateNet = private
	var (
		chain *vm_dummy.DummyChain
		err   error
	)
	if privateNet {
		chain, err = vm_dummy.LoadDummyChain()
	} else {
		chain, err = vm_dummy.LoadDummyChain(vm_dummy.SetPubNet())
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

func Get() *vm_dummy.DummyChain {
	if currentCtx == nil {
		log.Fatal("fail to open chain")
	}
	return currentCtx.chain
}
