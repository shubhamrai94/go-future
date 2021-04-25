package future

import (
	"reflect"
	"sync"
	"time"
	"fmt"
)

const _PENDING = "PENDING"
const _CANCELLED = "CANCELLED"
const _FINISHED = "FINISHED"

type State struct {
	value string
	mu    sync.Mutex
}

func (s *State) update(ns string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = ns

	return ns
}

func (s *State) get() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.value
}

type Future struct {
	functionType  reflect.Type
	functionValue reflect.Value
	args          []interface{}
	wait          chan bool
	result        interface{}
	err           error
	state         State
	callbacks     chan func(*Future)
}

type CancelledError string

func (e CancelledError) Error() string {
	return fmt.Sprintf("Error! %v", string(e))
}

type TimeOutError string

func (e TimeOutError) Error() string {
	return fmt.Sprintf("Error! %v", string(e))
}

func (f *Future) Cancel() bool {
	s := f.state.get()
	if s == _PENDING || s == _FINISHED {
		return false
	}

	f.state.update(_CANCELLED)

	return true
}

func (f *Future) Cancelled() bool {
	if f.state.get() == _CANCELLED {
		return true
	}

	return false
}

func (f *Future) Running() bool {
	if f.state.get() == _PENDING {
		return true
	}

	return false
}

func (f *Future) Done() bool {
	s := f.state.get()
	if s == _CANCELLED || s == _FINISHED {
		return true
	}

	return false
}

func (f *Future) Result(t time.Duration) (interface{}, error) {
	if f.Cancelled() {
		return nil, CancelledError("Future Cancelled")
	}
	if f.Done() {
		return f.result, f.err
	}

	if int(t) == 0 {
		<-f.wait

		return f.result, f.err
	}

	select {
	case <-f.wait:
		return f.result, f.err
	case <-time.After(t):
		return nil, TimeOutError("Time Out")
	}
}

func (f *Future) Exception(t time.Duration) (error) {
	if f.Cancelled() {
		return CancelledError("Future Cancelled")
	}
	if f.Done() {
		return f.err
	}

	if int(t) == 0 {
		<-f.wait

		return f.err
	}

	select {
	case <-f.wait:
		return f.err
	case <-time.After(t):
		return TimeOutError("Time Out")
	}
}

func (f *Future) AddDoneCallback(function func (*Future)) {
	go func () {
		f.callbacks <- function
	}()
}

func (f *Future) SetResult(r interface{}) {
	f.result = r
}

func (f *Future) run() {
	if f.Cancelled() {
		return
	}

	f.state.update(_PENDING)

	defer func () {
		f.state.update(_FINISHED)
		f.wait <- false

		go f.processCallbacks()
	}()

	numParams := f.functionType.NumIn()
	values := make([]reflect.Value, numParams)
	for i := 0; i < numParams; i++ {
		values[i] = reflect.ValueOf(f.args[i])
	}

	ret := f.functionValue.Call(values)

	if len(ret) == 0 {
		return
	}

	f.result = ret[0]

	if f.functionType.NumOut() > 1 && !ret[1].IsNil() {
		f.err = ret[1].Interface().(error)
	}
}

func (f *Future) processCallbacks() {
	for {
		v := <-f.callbacks
		v(f)
	}
}

type Callable func(args ...interface{}) *Future

func New(function interface{}) Callable {
	functionType := reflect.TypeOf(function)
	if functionType.Kind() != reflect.Func {
		return nil
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if functionType.NumOut() > 1 && !functionType.Out(1).Implements(errorInterface) {
		return nil
	}

	return func(args ...interface{}) *Future {
		future := &Future{
			functionType:  functionType,
			functionValue: reflect.ValueOf(function),
			args:          args,
			wait:          make(chan bool),
			callbacks:     make(chan func(*Future)),
		}

		go future.run()

		return future
	}
}

