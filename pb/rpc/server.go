package rpc

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"

	"github.com/xjdrew/daisy/gen/proto/base"
)

type Server struct {
	bridge     *Bridge
	mu         sync.RWMutex
	serviceMap map[int32]*service
}

func newServer(bridge *Bridge) *Server {
	return &Server{
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
func (server *Server) RegisterModule(receiver interface{}) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	rcvr := reflect.ValueOf(receiver)
	typ := reflect.TypeOf(receiver)
	if err := server.register(rcvr, typ); err != nil {
		return err
	}

	if err := server.register(rcvr, reflect.PtrTo(typ)); err != nil {
		return err
	}

	return nil
}

func (server *Server) register(rcvr reflect.Value, typ reflect.Type) error {
	module := reflect.Indirect(rcvr).Type().Name()
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)

		dptor := server.bridge.getDescriptor(module, method.Name)
		if dptor == nil {
			return fmt.Errorf("undefined method %s.%s", module, method.Name)
		}

		if err := dptor.MatchMethod(method); err != nil {
			return err
		}

		if _, present := server.serviceMap[dptor.Id]; present {
			return fmt.Errorf("repeated method %s", dptor.MethodName)
		}

		server.serviceMap[dptor.Id] = &service{
			rcvr:   rcvr,
			method: method,
			dptor:  dptor,
		}
	}
	return nil
}

func (server *Server) onIoError(context *Context, err error) {
	log.Println("onIoError:", context, err)
}

// return true if ignore
func (server *Server) onUnknownPack(context *Context, pack *proto_base.Pack) bool {
	log.Println("onUnknownPack:", context, pack)
	return false
}

func (server *Server) getService(typ int32) *service {
	server.mu.RLock()
	defer server.mu.RUnlock()
	s := server.serviceMap[typ]
	return s
}

func (server *Server) Accept(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					return err
				}
			}
			continue
		}

		context := NewContext(server, conn)
		go context.serve()
	}
}
