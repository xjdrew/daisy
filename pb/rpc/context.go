package rpc

import (
	"net"
	"reflect"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/proto/base"
)

type Context struct {
	server *Server
	conn   net.Conn
	codec  *Codec
}

func NewContext(server *Server, conn net.Conn) *Context {
	return &Context{
		server: server,
		conn:   conn,
		codec:  NewCodec(conn),
	}
}

func (c *Context) writePack(pack *proto_base.Pack) {
	err := c.codec.WritePack(pack)
	if err != nil {
		c.server.onIoError(c, err)
		c.Close()
	}
}

func (c *Context) dispatchResponse(pack *proto_base.Pack) bool {
	return true
}

func (c *Context) dispatchRequest(pack *proto_base.Pack) bool {
	typ := pack.GetType()
	s := c.server.getService(typ)
	if s == nil {
		return c.server.onUnknownPack(c, pack)
	}

	argv := reflect.New(s.dptor.ArgType.Elem())
	if err := proto.Unmarshal(pack.GetData(), argv.Interface().(proto.Message)); err != nil {
		return c.server.onUnknownPack(c, pack)
	}

	var replyv reflect.Value
	if s.hasReply() {
		replyv = reflect.New(s.dptor.ReplyType.Elem())
	}

	go func() {
		if s.hasReply() {
			err := s.call(c, argv, replyv)
			pack.Type = proto.Int32(0)
			pack.Data = nil
			if err != nil {
				pack.GetError().Error = proto.String(err.Error())
			} else {
				pack.Data, _ = proto.Marshal(replyv.Interface().(proto.Message))
			}
			c.writePack(pack)
		} else {
			s.invoke(c, argv)
		}
	}()
	return true
}

func (c *Context) Close() {
	c.codec.Close()
}

func (c *Context) serve() {
	for {
		var pack proto_base.Pack
		if err := c.codec.ReadPack(&pack); err != nil {
			c.server.onIoError(c, err)
			break
		}

		typ := pack.GetType()
		var keepServing bool
		if typ == 0 { // response
			keepServing = c.dispatchResponse(&pack)
		} else {
			keepServing = c.dispatchRequest(&pack)
		}
		if !keepServing {
			break
		}
	}
	c.Close()
}
