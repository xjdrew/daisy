package rpc

import (
	"io"
	"log"
)

type Bridge struct {
	nameMap   map[string]*Descriptor
	methodMap map[string]*Descriptor
	idMap     map[int32]*Descriptor
}

func NewBridge(descriptors []Descriptor) *Bridge {
	idMap := make(map[int32]*Descriptor)
	nameMap := make(map[string]*Descriptor)
	methodMap := make(map[string]*Descriptor)
	for _, dptor := range descriptors {
		if _, ok := idMap[dptor.Id]; ok {
			log.Panicf("repeated service:%#v", dptor)
		} else {
			idMap[dptor.Id] = &dptor
		}

		nameMap[dptor.NormalName] = &dptor
		methodMap[dptor.MethodName] = &dptor
	}

	return &Bridge{
		idMap:     idMap,
		nameMap:   nameMap,
		methodMap: methodMap,
	}
}

func (bridge *Bridge) NewServer() *Server {
	return newServer(bridge)
}

func (bridge *Bridge) Dail(network, address string) (*Client, error) {
	return nil, nil
}

func (bridge *Bridge) NewClient(conn io.ReadWriteCloser) *Client {
	return nil
}

func (bridge *Bridge) getDescriptor(module, method string) *Descriptor {
	if service, ok := bridge.methodMap[module+"."+method]; ok {
		return service
	} else {
		return nil
	}
}
