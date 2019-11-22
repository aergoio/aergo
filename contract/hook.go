// +build !Debug

package contract

/*
#include "vm.h"
*/
import "C"

func (ce *executor) setCountHook(limit C.int) {
	if ce == nil ||
		ce.L == nil ||
		ce.err != nil ||
		vmIsGasSystem(ce.ctx) {
		C.vm_set_timeout_hook(ce.L)
		return
	}
	if HardforkConfig.IsV2Fork(ce.ctx.blockInfo.No) {
		C.vm_set_timeout_count_hook(ce.L, limit)
	} else {
		C.vm_set_count_hook(ce.L, limit)
	}
}
