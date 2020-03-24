package contract

/*
#include <lualib.h>
#include "lgmp.h"
#include "vm.h"
*/
import "C"
import (
	"runtime"
	"sync"
)

var getCh chan *LState
var freeCh chan *LState
var once sync.Once

func StartLStateFactory(num int) {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		getCh = make(chan *LState, num)
		freeCh = make(chan *LState, num)

		for i := 0; i < num; i++ {
			getCh <- newLState()
		}
		for i := 0; i < runtime.NumCPU()/2; i++ {
			go statePool()
		}
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
