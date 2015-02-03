package rpc

import ()

type Client struct {
}

func (client *Client) Call(method string, args interface{}, replay interface{}) error {
	return nil
}

func (client *Client) Send(method string, args interface{}) {
}

func (client *Client) Close() error {
	return nil
}
