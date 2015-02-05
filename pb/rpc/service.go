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
	return s.dptor.ReplyType != nil
}

// have response
func (s *service) call(c *Context, argv, replyv reflect.Value) error {
	function := s.method.Func
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	return returnValues[0].Interface().(error)
}

// no response
func (s *service) invoke(c *Context, argv reflect.Value) {
	function := s.method.Func
	function.Call([]reflect.Value{s.rcvr, argv})
}
