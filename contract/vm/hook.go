//go:build !Debug
// +build !Debug

package main

/*
#include "vm.h"
*/
import "C"

func (ce *executor) setCountHook(limit C.int) {
	if ce == nil || ce.L == nil || ce.err != nil || IsGasSystem() {
		C.vm_set_timeout_hook(ce.L)
		return
	}
	if hardforkVersion >= 2 {
		C.vm_set_timeout_count_hook(ce.L, limit)
	} else {
		C.vm_set_count_hook(ce.L, limit)
	}
}
