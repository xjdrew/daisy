package main

import (
	"github.com/xjdrew/daisy/gen/proto/debug"
)

type Debug int

func (d *Debug) Ping(req *proto_debug.Ping, rsp *proto_debug.Ping_Response) error {
	rsp.Pong = req.Ping
	return nil
}
