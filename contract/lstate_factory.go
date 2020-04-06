package contract

/*
#include <lualib.h>
#include "lgmp.h"
#include "vm.h"
*/
import "C"
import (
	"sync"
)

var getCh chan *LState
var freeCh chan *LState
var once sync.Once

func StartLStateFactory(num, numClosers, numCloseLimit int) {

	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		getCh = make(chan *LState, num)
		freeCh = make(chan *LState, num)

		for i := 0; i < num; i++ {
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
		state := <-freeCh
		s.append(state)
		getCh <- newLState()
	}
}

func getLState() *LState {
	state := <-getCh
	return state
}

func freeLState(state *LState) {
	freeCh <- state
}
