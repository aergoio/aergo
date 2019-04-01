package chain

import (
	"fmt"
	"sync"
)

type ErrDebug struct {
	cond stopCond
}

type stopCond int

// stop before swap chain
const (
	DEBUG_REORG_STOP_1 stopCond = 1 + iota
	DEBUG_REORG_STOP_2
	DEBUG_REORG_STOP_3
)

var stopConds = [...]string{
	"DEBUG_REORG_STOP_1",
	"DEBUG_REORG_STOP_2",
	"DEBUG_REORG_STOP_3",
}

func (c stopCond) String() string { return stopConds[c] }

func (ec *ErrDebug) Error() string {
	return fmt.Sprintf("stopped by debugger cond[%s]", ec.cond.String())
}

type debugger struct {
	sync.RWMutex
	condMap map[stopCond]bool
}

func newDebugger() *debugger {
	return &debugger{condMap: make(map[stopCond]bool)}
}

func (debug *debugger) set(cond stopCond) {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	debug.condMap[cond] = true
}

func (debug *debugger) unset(cond stopCond) {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	delete(debug.condMap, cond)
}

func (debug *debugger) clear() {
	if debug == nil {
		return
	}

	debug.Lock()
	defer debug.Unlock()

	debug.condMap = make(map[stopCond]bool)
}

func (debug *debugger) check(cond stopCond) error {
	if debug == nil {
		return nil
	}

	debug.Lock()
	defer debug.Unlock()

	if _, ok := debug.condMap[cond]; ok {
		return &ErrDebug{cond: cond}
	}

	return nil
}
