package contract

var getCh chan *LState
var freeCh chan *LState

const MAX_LSTATE_SIZE = 100

func init() {
	getCh = make(chan *LState, MAX_LSTATE_SIZE)
	freeCh = make(chan *LState, MAX_LSTATE_SIZE)

	go stateCreator()
	go stateDestructor()
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
