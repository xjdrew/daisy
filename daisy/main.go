package main

import (
	"log"
	"net"

	"github.com/xjdrew/daisy/gen/descriptor"
	"github.com/xjdrew/daisy/pb/rpc"
)

func main() {
	bridge := rpc.NewBridge(descriptor.Descriptors)
	server := bridge.NewServer()
	server.RegisterModule(new(Debug))
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	server.Accept(l)
}
