/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
)

// ICompSyncRequester is the interface that wraps the RequestFuture method.
type ICompSyncRequester interface {
	RequestFuture(targetName string, message interface{}, timeout time.Duration, tip string) *actor.Future
}

type ComponentHub struct {
	components map[string]IComponent
}

type hubInitSync struct {
	sync.WaitGroup
	finished chan interface{}
}

var hubInit hubInitSync

func NewComponentHub() *ComponentHub {
	hub := ComponentHub{
		components: make(map[string]IComponent),
	}
	return &hub
}

func (h *hubInitSync) begin(n int) {
	h.finished = make(chan interface{})
	h.Add(n)
}

func (h *hubInitSync) end() {
	h.Wait()
	close(h.finished)
}

func (h *hubInitSync) wait() {
	h.Done()
	<-h.finished
}

func (hub *ComponentHub) Start() {
	hubInit.begin(len(hub.components))
	for _, comp := range hub.components {
		go comp.Start()
	}
	hubInit.end()
}

func (hub *ComponentHub) Stop() {
	for _, comp := range hub.components {
		comp.Stop()
	}
}

func (hub *ComponentHub) Register(component IComponent) {
	hub.components[component.GetName()] = component
	component.SetHub(hub)
}

func (hub *ComponentHub) Statistics(timeOutSec time.Duration) map[string]*CompStatRsp {
	var compStatus map[string]Status
	compStatus = make(map[string]Status)

	// check a status of all components before ask a profiling
	// request the profiling to only alive components
	for _, comp := range hub.components {
		compStatus[comp.GetName()] = comp.Status()
	}

	// get current time and add this to a request
	// to estimate standing time at an actor's mailbox
	msgQueuedTime := time.Now()

	var jobMap map[string]*actor.Future
	jobMap = make(map[string]*actor.Future)
	var retCompStatics map[string]*CompStatRsp
	retCompStatics = make(map[string]*CompStatRsp)

	for name, comp := range hub.components {
		if compStatus[name] == StartedStatus {
			// send a request to all component asynchronously
			jobMap[name] = comp.RequestFuture(
				&CompStatReq{msgQueuedTime},
				timeOutSec,
				"pkg/component/hub.Status")
		} else {
			// in the case of non-started components, just record its status
			retCompStatics[name] = &CompStatRsp{
				Status: StatusToString(compStatus[name]),
			}
		}
	}

	// for each asynchronously thrown jobs
	for name, job := range jobMap {
		// wait and get a result
		result, err := job.Result()
		if err != nil {
			// when error is occurred, record it.
			// the most frequently occurred error will be a timeout error
			retCompStatics[name] = &CompStatRsp{
				Status: StatusToString(compStatus[name]),
				Error:  err.Error(),
			}
		} else {
			// in normal case, success, record response
			retCompStatics[name] = result.(*CompStatRsp)
		}
	}

	return retCompStatics
}
func (hub *ComponentHub) Tell(targetName string, message interface{}) {
	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		panic("Unregistered Component")
	}

	targetComponent.Tell(message)
}

func (hub *ComponentHub) RequestFuture(
	targetName string, message interface{}, timeout time.Duration, tip string) *actor.Future {

	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		panic("Unregistered Component")
	}

	return targetComponent.RequestFuture(message, timeout, tip)
}

func (hub *ComponentHub) Get(targetName string) IComponent {
	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		panic("Unregistered Component")
	}

	return targetComponent
}
