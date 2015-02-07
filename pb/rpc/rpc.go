/*
	Rpc struct discribe common things between Client and Rpc
*/
package rpc

import (
	"fmt"
	"reflect"
	"sync"
)

type Rpc struct {
	mu         sync.RWMutex
	serviceMap map[int32]*service
	bridge     *Bridge
}

func NewRpc(bridge *Bridge) Rpc {
	return Rpc{
		bridge:     bridge,
		serviceMap: make(map[int32]*service),
	}
}

/*
	注册模块
	模块中的函数必须满足以下几点才能成为接口函数
	1. defined in protolist
	2. 函数原型满足protolist里面对应接口的定义
	如果还有其他类型的函数，则返回error
*/
func (r *Rpc) RegisterModule(receiver interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	typ := reflect.TypeOf(receiver)
	rcvr := reflect.ValueOf(receiver)

	if err := r.register(rcvr, typ); err != nil {
		return err
	}

	if err := r.register(rcvr, reflect.PtrTo(typ)); err != nil {
		return err
	}

	return nil
}

func (r *Rpc) register(rcvr reflect.Value, typ reflect.Type) error {
	module := reflect.Indirect(rcvr).Type().Name()
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)

		dptor := r.bridge.getDescriptor(module, method.Name)
		if dptor == nil {
			return fmt.Errorf("undefined method %s.%s", module, method.Name)
		}

		if err := dptor.MatchMethod(method); err != nil {
			return err
		}

		if _, present := r.serviceMap[dptor.Id]; present {
			return fmt.Errorf("repeated method %s", dptor.MethodName)
		}

		r.serviceMap[dptor.Id] = &service{
			rcvr:   rcvr,
			method: method,
			dptor:  dptor,
		}
	}
	return nil
}

func (r *Rpc) getService(typ int32) *service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s := r.serviceMap[typ]
	return s
}

func (r *Rpc) getDescriptor(name string) *Descriptor {
	return r.bridge.getDescriptorByName(name)
}
