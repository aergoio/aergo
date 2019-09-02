package contract

/*
#include <lualib.h>
#include "lgmp.h"
#include "vm.h"
*/
import "C"
import "sync"

var getCh chan *LState
var freeCh chan *LState
var once sync.Once

const lStateMaxSize = 150

func StartLStateFactory() {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		getCh = make(chan *LState, lStateMaxSize)
		freeCh = make(chan *LState, lStateMaxSize)

		for i := 0; i < lStateMaxSize; i++ {
			getCh <- newLState()
		}
		go statePool()
	})
}

func statePool() {
	for {
		state := <-freeCh
		state.close()
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
