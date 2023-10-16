package rpc

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/pkg/component"
)

func TestActorTimeout(t *testing.T) {
	logger := log.NewLogger("rpc.test")
	compMng := component.NewComponentHub()

	a := &AActor{}
	a.BaseComponent = component.NewBaseComponent("a", a, logger)
	compMng.Register(a)
	compMng.Start()
	defer compMng.Stop()

	var f *actor.Future
	// succ
	f = compMng.RequestFuture("a", TimedMsg{}, time.Second>>2, "nop")
	if r, err := f.Result(); err != nil {
		t.Errorf("want no error, but not. result is %v", err.Error())
	} else {
		resp, ok := r.(TimedMsgRsp)
		if !ok {
			t.Errorf("want result type %v, but %v", reflect.TypeOf(TimedMsgRsp{}), reflect.TypeOf(resp))
		} else {
			if resp.Err != nil {
				t.Errorf("want succ, but not %s", resp.Err.Error())
			}
		}
	}

	// panic
	f = compMng.RequestFuture("a", TimedMsg{p: true}, time.Second>>2, "nop")
	if r, err := f.Result(); err == nil {
		t.Errorf("want timeout, but not. result is %v", reflect.TypeOf(r))
		rErr := r.(error)
		t.Logf("err in result: %v", rErr.Error())
	} else {
		t.Logf("Err type %v, err msg %s", reflect.TypeOf(err), err.Error())
	}

	// failure
	f = compMng.RequestFuture("a", TimedMsg{e: true}, time.Second>>2, "nop")
	if r, err := f.Result(); err != nil {
		t.Errorf("want no error, but not. result is %v", err.Error())
	} else {
		resp, ok := r.(TimedMsgRsp)
		if !ok {
			t.Errorf("want result type %v, but %v", reflect.TypeOf(TimedMsgRsp{}), reflect.TypeOf(resp))
		} else {
			if resp.Err == nil {
				t.Errorf("want failed, but not %s", resp.Err.Error())
			}
		}
	}
	// timeout
	f = compMng.RequestFuture("a", TimedMsg{wt: time.Minute}, time.Second>>2, "nop")
	f2 := compMng.RequestFuture("a", TimedMsg{wt: 0}, time.Second>>2, "nop")
	f3 := compMng.RequestFuture("a", TimedMsg{wt: time.Second >> 3}, time.Second>>2, "nop")
	if r, err := f.Result(); err == nil {
		t.Errorf("want timeout, but not. result is %v", reflect.TypeOf(r))
		rErr := r.(error)
		t.Logf("err in result: %v", rErr.Error())
	} else if actor.ErrTimeout != err {
		t.Logf("Err %v, want global var actor.ErrTimeout", reflect.TypeOf(err))
	} else {
		t.Logf("Err type %v, err msg %s", reflect.TypeOf(err), err.Error())
	}
	r, err := f2.Result()
	t.Logf("pended first future is : %v, err %v", r, err)
	r, err = f3.Result()
	t.Logf("pended second future is : %v, err %v", r, err)
}

type TimedMsg struct {
	wt time.Duration
	e  bool
	p  bool
}
type TimedMsgRsp struct {
	Err error
}

type AActor struct {
	*component.BaseComponent
}

func (a *AActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case TimedMsg:
		if msg.wt > 0 {
			time.Sleep(msg.wt)
		}
		if msg.p {
			panic("it's panic")
		}
		var err error
		if msg.e {
			err = errors.New("error while executing actor message")
		}
		context.Respond(TimedMsgRsp{Err: err})
	}
}
func (a *AActor) BeforeStart() {}
func (a *AActor) AfterStart()  {}
func (a *AActor) BeforeStop()  {}

func (a *AActor) Statistics() *map[string]interface{} {
	stat := make(map[string]interface{})
	return &stat
}
