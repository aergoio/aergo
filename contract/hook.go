//go:build !Debug
// +build !Debug

package contract

/*
#include "vm.h"
*/
import "C"

import "github.com/aergoio/aergo/v2/fee"

func (ce *executor) setCountHook(limit C.int) {
	if ce == nil ||
		ce.L == nil ||
		ce.err != nil ||
		fee.IsVmGasSystem(ce.ctx.blockInfo.ForkVersion, ce.ctx.isQuery) {
		C.vm_set_timeout_hook(ce.L)
		return
	}
	if ce.ctx.blockInfo.ForkVersion >= 2 {
		C.vm_set_timeout_count_hook(ce.L, limit)
	} else {
		C.vm_set_count_hook(ce.L, limit)
	}
}
