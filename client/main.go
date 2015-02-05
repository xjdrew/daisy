package main

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/xjdrew/daisy/gen/descriptor"
	"github.com/xjdrew/daisy/gen/proto/debug"
	"github.com/xjdrew/daisy/pb/rpc"
)

func main() {
	bridge := rpc.NewBridge(descriptor.Descriptors)
	client, err := bridge.Dail("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()

	ping := proto_debug.Ping{
		Ping: proto.String("hello"),
	}
	pong := proto_debug.Ping_Response{}
	err = client.Call("debug.ping", ping, &pong)
	if err != nil {
		log.Fatal("call debug.ping:", err)
	}
	log.Printf("debug.ping: %s -> %s", ping.Ping, pong.Pong)
}
