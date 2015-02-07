package rpc

import (
	"reflect"
)

type service struct {
	rcvr   reflect.Value
	method reflect.Method
	dptor  *Descriptor
}

func (s *service) hasReply() bool {
	return s.dptor.HasReply()
}

// have response
func (s *service) call(c *Context, argv, replyv reflect.Value) *CallError {
	function := s.method.Func
	returnValues := function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(c), argv, replyv})

	inter := returnValues[0].Interface()
	if inter == nil {
		return nil
	}
	return inter.(*CallError)
}

// no response
func (s *service) invoke(c *Context, argv reflect.Value) {
	function := s.method.Func
	function.Call([]reflect.Value{s.rcvr, reflect.ValueOf(c), argv})
}
