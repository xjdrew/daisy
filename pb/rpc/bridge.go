package rpc

import (
	"log"
	"net"
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
	for i := range descriptors {
		dptor := descriptors[i]
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

func (bridge *Bridge) Dail(network, address string) (cli *Client, err error) {
	var conn net.Conn
	conn, err = net.Dial(network, address)
	if err != nil {
		return
	}
	cli = bridge.NewClient(conn)
	return
}

func (bridge *Bridge) NewClient(conn net.Conn) *Client {
	return NewClient(bridge, conn)
}

func (bridge *Bridge) getDescriptorByName(name string) *Descriptor {
	return bridge.nameMap[name]
}

func (bridge *Bridge) getDescriptor(module, method string) *Descriptor {
	return bridge.methodMap[module+"."+method]
}
