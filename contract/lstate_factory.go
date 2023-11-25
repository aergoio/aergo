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
	ctrLgr.Debug().Msg("LState acquired")
	return state
}

func FreeLState(state *LState) {
	if state != nil {
		freeCh <- state
		ctrLgr.Debug().Msg("LState released")
	}
}

func FlushLStates() {
	for i := 0; i < maxLStates; i++ {
		s := GetLState()
		FreeLState(s)
	}
}
