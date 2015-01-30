package main

import (
	"log"
	"reflect"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/interfaces"
	"github.com/xjdrew/daisy/gen/proto/debug"
)

func main() {
	ping := &proto_debug.Ping{
		Ping: proto.String("hello"),
	}

	pong := &proto_debug.Ping_Response{}
	log.Printf("%v", reflect.TypeOf(ping).Elem().Name())
	log.Printf("%s", reflect.TypeOf(pong).String())
	data, err := proto.Marshal(ping)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	newPing := &proto_debug.Ping{}
	if err = proto.Unmarshal(data, newPing); err != nil {
		log.Fatal("unmarshaling error:", err)
	}

	if ping.GetPing() != newPing.GetPing() {
		log.Fatalf("data mismatch %q != %q", ping.GetPing(), newPing.GetPing())
	}

	log.Printf("Modules count: %d", len(protolist.Modules))
}
