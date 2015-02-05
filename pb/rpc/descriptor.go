package rpc

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

type Descriptor struct {
	Id         int32
	NormalName string
	MethodName string
	ArgType    reflect.Type
	ReplyType  reflect.Type
}

var typeOfProtoMessage = reflect.TypeOf((*proto.Message)(nil)).Elem()
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

/*
	如果接口有返回值，则需要两个参数，类型为指针，有一个error类型的返回值
	如果接口没有返回值，则只需要一个参数，类型为指针， 没有返回值
*/
func (d *Descriptor) MatchMethod(method reflect.Method) error {
	mtyp := method.Type
	numIn := 2
	numOut := 0
	if d.ReplyType != nil {
		numIn += 1
		numOut += 1
	}

	if mtyp.NumIn() != numIn {
		return fmt.Errorf("method %s should have %d arguments", d.MethodName, numIn)
	}

	if mtyp.NumOut() != numOut {
		return fmt.Errorf("method %s should have %d return values", d.MethodName, numOut)
	}

	argType := mtyp.In(1)
	if argType.Kind() != reflect.Ptr || !argType.Implements(typeOfProtoMessage) || argType != d.ArgType {
		return fmt.Errorf("method %s arg%d should be %s", d.MethodName, 1, d.ArgType.String())
	}

	if numIn > 2 {
		replyType := mtyp.In(2)
		if replyType.Kind() != reflect.Ptr || !replyType.Implements(typeOfProtoMessage) || replyType != d.ReplyType {
			return fmt.Errorf("method %s arg%d should be %s", d.MethodName, 2, d.ReplyType.String())
		}
	}

	if numOut > 0 {
		if returnType := mtyp.Out(0); returnType != typeOfError {
			return fmt.Errorf("method %s returns %s not error", d.MethodName, returnType.String())
		}
	}
	return nil
}
