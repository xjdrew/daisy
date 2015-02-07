package main

import (
	"log"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/descriptor"
	"github.com/xjdrew/daisy/gen/proto/test"
	"github.com/xjdrew/daisy/pb/rpc"
)

type Test int

func (t *Test) Echo(context *rpc.Context, req *proto_test.Echo, rsp *proto_test.Echo_Response) *rpc.CallError {
	log.Printf("Echo: %+v", req)
	context.MustInvoke("test.strobe", &proto_test.Strobe{Msg: proto.String("recv:" + req.GetReq())})
	rsp.Resp = req.Req
	return nil
}

func register(server *rpc.Server, rcvr interface{}) {
	err := server.RegisterModule(rcvr)
	if err != nil {
		log.Fatal("register error:", err)
	}
}

func main() {
	bridge := rpc.NewBridge(descriptor.Descriptors)
	server := bridge.NewServer()
	register(server, new(Debug))
	register(server, new(Test))
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	server.Accept(l)
}
