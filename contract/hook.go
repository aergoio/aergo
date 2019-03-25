// +build !Debug

package contract

/*
#include "vm.h"
*/
import "C"

func (ce *Executor) setCountHook(limit C.int) {
	if ce == nil || ce.L == nil {
		return
	}
	if ce.err != nil {
		return
	}
	C.vm_set_count_hook(ce.L, limit)
}
