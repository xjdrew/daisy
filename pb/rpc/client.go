package rpc

import (
	"log"
	"net"

	"github.com/xjdrew/daisy/gen/proto/base"
)

type Client struct {
	*Rpc
	*Context
}

func NewClient(bridge *Bridge, conn net.Conn) *Client {
	cli := new(Client)
	cli.Rpc = &Rpc{bridge: bridge, serviceMap: make(map[int32]*service)}
	cli.Context = NewContext(cli, conn)
	return cli
}

func (client *Client) onIoError(context *Context, err error) {
	log.Println("onIoError:", context, err)
}

// return true if ignore
func (client *Client) onUnknownPack(context *Context, pack *proto_base.Pack) bool {
	log.Println("onUnknownPack:", context, pack)
	return true
}

func (client *Client) Serve() {
	client.Context.serve()
}

func (client *Client) Close() error {
	return client.Context.Close()
}
