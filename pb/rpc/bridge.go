package rpc

import (
	"io"

	"github.com/xjdrew/daisy/pb/parser"
)

type Bridge struct {
}

func NewBridge(modules []parser.Module) *Bridge {
	return nil
}

func (bridge *Bridge) NewServer() *Server {
	return nil
}

func (bridge *Bridge) Dail(network, address string) (*Client, error) {
	return nil, nil
}

func (bridge *Bridge) NewClient(conn io.ReadWriteCloser) *Client {
	return nil
}
