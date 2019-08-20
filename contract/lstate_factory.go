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

const MAX_LSTATE_SIZE = 150

func StartLStateFactory() {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		getCh = make(chan *LState, MAX_LSTATE_SIZE)
		freeCh = make(chan *LState, MAX_LSTATE_SIZE)

		for i := 0; i < MAX_LSTATE_SIZE; i++ {
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

func GetLState() *LState {
	state := <-getCh
	return state
}

func FreeLState(state *LState) {
	freeCh <- state
}
