package rpc

import (
	"net"
)

type Server struct {
}

func (server *Server) RegisterModule(rcvr interface{}) error {
	return nil
}

func (server *Server) Accept(lis net.Listener) {
}
