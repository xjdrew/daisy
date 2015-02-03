package rpc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/proto/base"
)

const (
	BUFLEN = 65535
)

type Codec struct {
	rwc   io.ReadWriteCloser
	rdbuf [BUFLEN]byte
}

func (c *Codec) ReadPack(p *proto_base.Pack) error {
	if p == nil {
		return nil
	}

	var sz uint16
	if err := binary.Read(c.rwc, binary.BigEndian, &sz); err != nil {
		return err
	}

	var to uint16 = 0
	for to < sz {
		n, err := c.rwc.Read(c.rdbuf[to:])
		if err != nil {
			return err
		}
		to += uint16(n)
	}

	if err := proto.Unmarshal(c.rdbuf[:sz], p); err != nil {
		return err
	}
	return nil
}

func (c *Codec) WritePack(p *proto_base.Pack) error {
	if p == nil {
		return nil
	}

	data, err := proto.Marshal(p)
	if err != nil {
		return err
	}

	sz := len(data)
	if sz > BUFLEN {
		return fmt.Errorf("WritePack: overflow packet size(%d)", sz)
	}

	if err = binary.Write(c.rwc, binary.BigEndian, sz); err != nil {
		return err
	}

	if _, err = c.rwc.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Codec) Close() error {
	return c.rwc.Close()
}
