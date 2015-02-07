package main

import (
	"log"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/descriptor"
	"github.com/xjdrew/daisy/gen/proto/test"
	"github.com/xjdrew/daisy/pb/rpc"
)

type Test int

func (t *Test) Strobe(context *rpc.Context, req *proto_test.Strobe) {
	log.Printf("Strobe: %+v", req)
}

func main() {
	bridge := rpc.NewBridge(descriptor.Descriptors)
	client, err := bridge.Dail("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	// register callbacks
	if err := client.RegisterModule(new(Test)); err != nil {
		log.Fatal("register error:", err)
	}

	go client.Serve()
	// echo
	req := &proto_test.Echo{Req: proto.String("hello")}
	echod := proto_test.Echo_Response{}
	callErr := client.MustCall("test.echo", req, &echod)
	if callErr != nil {
		log.Fatal("call test.echo:", callErr)
	}
	log.Printf("echo response:%s", echod.GetResp())
}
