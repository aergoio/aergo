// +build !Debug

package contract

/*
#include "vm.h"
*/
import "C"

func (ce *Executor) setCountHook(limit C.int) {
	if ce == nil ||
		ce.L == nil ||
		ce.err != nil ||
		!vmNeedResourceLimit(ce.stateSet) {
		return
	}
	C.vm_set_count_hook(ce.L, limit)
}
