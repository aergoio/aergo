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
		return
	}
	C.vm_set_count_hook(ce.L, limit)
}
