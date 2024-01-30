package contract

/*
#include <lualib.h>
#include "bignum_module.h"
#include "vm.h"
*/
import "C"
import (
	"sync"
)

var maxLStates int
var getCh chan *LState
var freeCh chan *LState
var once sync.Once

func StartLStateFactory(numLStates, numClosers, numCloseLimit int) {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()

		maxLStates = numLStates
		getCh = make(chan *LState, numLStates)
		freeCh = make(chan *LState, numLStates)

		for i := 0; i < numLStates; i++ {
			getCh <- newLState()
		}

		for i := 0; i < numClosers; i++ {
			go statePool(numCloseLimit)
		}
	})
}

func statePool(numCloseLimit int) {
	s := newLStatesBuffer(numCloseLimit)

	for {
		select {
		case state := <-freeCh:
			s.append(state)
			getCh <- newLState()
		}
	}
}

func GetLState() *LState {
	state := <-getCh
	ctrLgr.Trace().Msg("LState acquired")
	return state
}

func FreeLState(state *LState) {
	if state != nil {
		freeCh <- state
		ctrLgr.Trace().Msg("LState released")
	}
}

func FlushLStates() {
	for i := 0; i < maxLStates; i++ {
		s := GetLState()
		FreeLState(s)
	}
}

//--------------------------------------------------------------------//
// LState type

type LState = C.struct_lua_State

func newLState() *LState {
	ctrLgr.Trace().Msg("LState created")
	return C.vm_newstate(C.int(currentForkVersion))
}

func (L *LState) close() {
	if L != nil {
		C.lua_close(L)
	}
}

type lStatesBuffer struct {
	s     []*LState
	limit int
}

func newLStatesBuffer(limit int) *lStatesBuffer {
	return &lStatesBuffer{
		s:     make([]*LState, 0),
		limit: limit,
	}
}

func (Ls *lStatesBuffer) len() int {
	return len(Ls.s)
}

func (Ls *lStatesBuffer) append(s *LState) {
	Ls.s = append(Ls.s, s)
	if Ls.len() == Ls.limit {
		Ls.close()
	}
}

func (Ls *lStatesBuffer) close() {
	C.vm_closestates(&Ls.s[0], C.int(len(Ls.s)))
	Ls.s = Ls.s[:0]
}
