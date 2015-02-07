package rpc

import (
	"log"
	"net"

	"github.com/xjdrew/daisy/gen/proto/base"
)

type Server struct {
	Rpc
}

func newServer(bridge *Bridge) *Server {
	return &Server{
		Rpc: Rpc{bridge: bridge, serviceMap: make(map[int32]*service)},
	}
}

func (r *Rpc) onIoError(context *Context, err error) {
	log.Println("onIoError:", context, err)
}

// return true if ignore
func (r *Rpc) onUnknownPack(context *Context, pack *proto_base.Pack) bool {
	log.Println("onUnknownPack:", context, pack)
	return false
}

func (server *Server) Accept(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					return err
				}
			}
			continue
		}

		context := NewContext(server, conn)
		go context.serve()
	}
}
