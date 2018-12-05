/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"errors"
	"sync"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/opentracing/opentracing-go"
	"github.com/satori/go.uuid"
)

var (
	ErrHubUnregistered = errors.New("Unregistered Component")
	logger             = log.NewLogger("actor")
)

// ICompSyncRequester is the interface that wraps the RequestFuture method.
type ICompSyncRequester interface {
	RequestFuture(targetName string, message interface{}, timeout time.Duration, tip string) *actor.Future
}

// ComponentHub keeps a list of registered components
type ComponentHub struct {
	components map[string]IComponent
	spanLock   sync.Mutex
	spans      map[string]*opentracing.Span
}

type hubInitSync struct {
	sync.WaitGroup
	finished chan interface{}
}

var hubInit hubInitSync

// NewComponentHub creates and returns ComponentHub instance
func NewComponentHub() *ComponentHub {
	hub := ComponentHub{
		components: make(map[string]IComponent),
		spans:      make(map[string]*opentracing.Span),
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

func (hub *ComponentHub) SaveSpan(span opentracing.Span) string {
	id := uuid.Must(uuid.NewV4()).String()
	hub.spanLock.Lock()
	defer hub.spanLock.Unlock()
	hub.spans[id] = &span
	return id
}

func (hub *ComponentHub) RestoreSpan(id string) *opentracing.Span {
	hub.spanLock.Lock()
	defer hub.spanLock.Unlock()
	return hub.spans[id]
}

func (hub *ComponentHub) DestroySpan(id string) {
	hub.spanLock.Lock()
	defer hub.spanLock.Unlock()
	span := hub.spans[id]
	if nil != span {
		(*span).Finish()
		//delete(hub.spans, id)
	}
}

// Start invokes start funcs of registered components at this hub
func (hub *ComponentHub) Start() {
	hubInit.begin(len(hub.components))
	for _, comp := range hub.components {
		go comp.Start()
	}
	hubInit.end()
}

// Stop invokes stop funcs of registered components at this hub
func (hub *ComponentHub) Stop() {
	for _, comp := range hub.components {
		comp.Stop()
	}
}

// Register assigns a component to this hub for management
func (hub *ComponentHub) Register(components ...IComponent) {
	for _, component := range components {
		if component != nil {
			hub.components[component.GetName()] = component
			component.SetHub(hub)
		}
	}
}

// Statistics invoke requests to all registered components,
// collect and return it's response
// An input argument, timeout, is used to set actor request's timeout.
// If it is over, than future: timeout string set at error field
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
	var retCompStatistics map[string]*CompStatRsp
	retCompStatistics = make(map[string]*CompStatRsp)

	for name, comp := range hub.components {
		if compStatus[name] == StartedStatus {
			// send a request to all component asynchronously
			jobMap[name] = comp.RequestFuture(
				&CompStatReq{msgQueuedTime},
				timeOutSec,
				"pkg/component/hub.Status")
		} else {
			// in the case of non-started components, just record its status
			retCompStatistics[name] = &CompStatRsp{
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
			retCompStatistics[name] = &CompStatRsp{
				Status:      StatusToString(compStatus[name]),
				MsgQueueLen: uint64(hub.Get(name).MsgQueueLen()),
				Error:       err.Error(),
			}
		} else {
			// in normal case, success, record response
			retCompStatistics[name] = result.(*CompStatRsp)
		}
	}

	return retCompStatistics
}

// Tell pass and forget a message to a component, which has a targetName
func (hub *ComponentHub) Tell(targetName string, message interface{}) {
	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		panic("Unregistered Component")
	}

	targetComponent.Tell(message)
}

// RequestFuture pass a message to a component, which has a targetName
// And this returns a future instance to be used in waiting a response
func (hub *ComponentHub) RequestFuture(
	targetName string, message interface{}, timeout time.Duration, tip string) *actor.Future {

	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		err := actor.NewFuture(timeout)
		err.PID().Tell(ErrHubUnregistered)
		return err
	}

	return targetComponent.RequestFuture(message, timeout, tip)
}

func (hub *ComponentHub) RequestFutureResult(
	targetName string, message interface{}, timeout time.Duration, tip string) (interface{}, error) {

	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		return nil, ErrHubUnregistered
	}

	return targetComponent.RequestFuture(message, timeout, tip).Result()
}

// Get returns a component which has a targetName
func (hub *ComponentHub) Get(targetName string) IComponent {
	targetComponent := hub.components[targetName]
	if targetComponent == nil {
		panic("Unregistered Component")
	}

	return targetComponent
}
