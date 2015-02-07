package rpc

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type CallError struct {
	Code     int32
	Msg      string
	RpcError bool // is rpc error
}

func NewCallError(code int32, format string, a ...interface{}) *CallError {
	err := new(CallError)
	err.Code = code
	err.Msg = fmt.Sprintf(format, a...)
	err.RpcError = false
	return err
}

func NewRpcCallError(code int32, format string, a ...interface{}) *CallError {
	err := new(CallError)
	err.Code = code
	err.Msg = fmt.Sprintf(format, a...)
	err.RpcError = true
	return err
}

func (e *CallError) String() string {
	if e == nil {
		return "<nil>"
	}
	var tag string
	if e.RpcError {
		tag = "rpc"
	} else {
		tag = "local"
	}
	return fmt.Sprintf("%s error: code:%d, msg:%s", tag, e.Code, e.Msg)
}

func (e *CallError) IsRpcError() bool {
	if e == nil {
		return false
	}
	return e.RpcError
}

type Descriptor struct {
	Id         int32
	NormalName string
	MethodName string
	ArgType    reflect.Type
	ReplyType  reflect.Type
}

var typeOfProtoMessage = reflect.TypeOf((*proto.Message)(nil)).Elem()
var typeOfContext = reflect.TypeOf(&Context{})
var typeOfError = reflect.TypeOf(&CallError{})

func (d *Descriptor) HasReply() bool {
	return d.ReplyType != nil
}

func (d *Descriptor) MatchArgType(typ reflect.Type) bool {
	if typ.Kind() != reflect.Ptr || !typ.Implements(typeOfProtoMessage) || typ != d.ArgType {
		return false
	}
	return true
}

func (d *Descriptor) MatchReplyType(typ reflect.Type) bool {
	if typ.Kind() != reflect.Ptr || !typ.Implements(typeOfProtoMessage) || typ != d.ReplyType {
		return false
	}
	return true
}

/*
	如果接口有返回值，则需要两个参数，类型为指针，有一个error类型的返回值
	如果接口没有返回值，则只需要一个参数，类型为指针， 没有返回值
*/
func (d *Descriptor) MatchMethod(method reflect.Method) error {
	mtyp := method.Type

	// defautl args: rcvr, context, req
	numIn := 3
	numOut := 0
	if d.ReplyType != nil {
		// extra arg: reply
		numIn += 1
		// extra return: *CallError
		numOut += 1
	}

	if mtyp.NumIn() != numIn {
		return fmt.Errorf("method %s should have %d arguments", d.MethodName, numIn)
	}

	if mtyp.NumOut() != numOut {
		return fmt.Errorf("method %s should have %d return values", d.MethodName, numOut)
	}

	contextType := mtyp.In(1)
	if contextType != typeOfContext {
		return fmt.Errorf("method %s arg%d should be %s", d.MethodName, 1, typeOfContext.String())
	}

	argType := mtyp.In(2)
	if !d.MatchArgType(argType) {
		return fmt.Errorf("method %s arg%d should be %s", d.MethodName, 2, d.ArgType.String())
	}

	if numIn > 3 {
		replyType := mtyp.In(3)
		if !d.MatchReplyType(replyType) {
			return fmt.Errorf("method %s arg%d should be %s", d.MethodName, 3, d.ReplyType.String())
		}
	}

	if numOut > 0 {
		if returnType := mtyp.Out(0); returnType != typeOfError {
			return fmt.Errorf("method %s returns %s not %s", d.MethodName, returnType.String(), typeOfError.String())
		}
	}
	return nil
}
