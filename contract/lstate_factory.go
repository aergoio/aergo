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

const maxLStateSize = 150

func StartLStateFactory() {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		getCh = make(chan *LState, maxLStateSize)
		freeCh = make(chan *LState, maxLStateSize)

		for i := 0; i < maxLStateSize; i++ {
			getCh <- NewLState()
		}
		go statePool()
	})
}

func statePool() {
	for {
		state := <-freeCh
		state.Close()
		getCh <- NewLState()
	}
}

func getLState() *LState {
	state := <-getCh
	return state
}

func freeLState(state *LState) {
	freeCh <- state
}
