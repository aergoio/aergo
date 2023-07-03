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

const (
	LStateDefault = iota
	LStateVer3
	LStateMax
)

var getCh [LStateMax]chan *LState
var freeCh [LStateMax]chan *LState
var once sync.Once

func StartLStateFactory(num, numClosers, numCloseLimit int) {
	once.Do(func() {
		C.init_bignum()
		C.initViewFunction()
		for i := 0; i < LStateMax; i++ {
			getCh[i] = make(chan *LState, num)
			freeCh[i] = make(chan *LState, num)
		}

		for i := 0; i < num; i++ {
			for j := 0; j < LStateMax; j++ {
				getCh[j] <- newLState(j)
			}
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
		case state := <-freeCh[LStateDefault]:
			s.append(state)
			getCh[LStateDefault] <- newLState(LStateDefault)
		case state := <-freeCh[LStateVer3]:
			s.append(state)
			getCh[LStateVer3] <- newLState(LStateVer3)
		}
	}
}

func GetLState(lsType int) *LState {
	state := <-getCh[lsType]
	ctrLgr.Debug().Int("type", lsType).Msg("LState acquired")
	return state
}

func FreeLState(state *LState, lsType int) {
	if state != nil {
		freeCh[lsType] <- state
		ctrLgr.Debug().Int("type", lsType).Msg("LState released")
	}
}
