package main

import (
	"log"

	"github.com/xjdrew/daisy/pb/rpc"

	"github.com/xjdrew/daisy/gen/proto/debug"
)

type Debug int

func (d *Debug) Ping(context *rpc.Context, req *proto_debug.Ping, rsp *proto_debug.Ping_Response) *rpc.CallError {
	log.Printf("Ping: %+v", req)
	rsp.Pong = req.Ping
	return nil
}
