package main

import (
	"log"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/pb/debug"
)

func main() {
	ping := &proto_debug.Ping{
		Ping: proto.String("hello"),
	}

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
}
