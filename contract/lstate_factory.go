package contract

/*
#include "lbc.h"
 */
import "C"
import "sync"

var getCh chan *LState
var freeCh chan *LState
var once sync.Once

const MAX_LSTATE_SIZE = 100

func StartLStateFactory() {
	once.Do(func() {
		C.bc_init_numbers()
		getCh = make(chan *LState, MAX_LSTATE_SIZE)
		freeCh = make(chan *LState, MAX_LSTATE_SIZE)

		go stateCreator()
		go stateDestructor()
	})
}

func stateCreator() {
	for {
		getCh <- NewLState()
	}
}

func stateDestructor() {
	for {
		state := <-freeCh
		state.Close()
	}
}

func GetLState() *LState {
	state := <-getCh
	return state
}

func FreeLState(state *LState) {
	freeCh <- state
}
