/*
	one connection one context
*/
package rpc

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/golang/protobuf/proto"

	"github.com/xjdrew/daisy/gen/proto/base"
)

type ContextOwner interface {
	getDescriptor(name string) *Descriptor
	getService(int32) *service
	onIoError(*Context, error)
	onUnknownPack(*Context, *proto_base.Pack) bool
}

type Call struct {
	Dptor *Descriptor
	Argv  interface{}
	Reply interface{}
	Error *CallError
	Done  chan *Call
}

func (call *Call) done() {
	select {
	case call.Done <- call:
	default:
		// done channel is unbuffered, it's a error
		// it's a error
		log.Panicf("method %s block", call.Dptor.NormalName)
	}
}

type Context struct {
	owner    ContextOwner
	conn     net.Conn
	codec    *Codec
	session  int32
	sessLock sync.Mutex
	sessions map[int32]*Call

	// err context
	err *error
}

func NewContext(owner ContextOwner, conn net.Conn) *Context {
	return &Context{
		owner:    owner,
		conn:     conn,
		codec:    NewCodec(conn),
		sessions: make(map[int32]*Call),
	}
}

func (c *Context) nextSession() int32 {
	return atomic.AddInt32(&c.session, 1)
}

func (c *Context) setSession(session int32, call *Call) {
	c.sessLock.Lock()
	c.sessions[session] = call
	c.sessLock.Unlock()
}

func (c *Context) grabSession(session int32) *Call {
	c.sessLock.Lock()
	defer c.sessLock.Unlock()
	if call, ok := c.sessions[session]; ok {
		delete(c.sessions, session)
		return call
	}
	return nil
}

func (c *Context) closeAllSessions() {
	c.sessLock.Lock()
	defer c.sessLock.Unlock()

	msg := "connection down"
	if c.err != nil {
		msg += ": " + (*c.err).Error()
	}

	for k, call := range c.sessions {
		delete(c.sessions, k)

		call.Error = NewCallError(0, msg)
		call.done()
	}
}

func (c *Context) setError(err error) {
	atomic.CompareAndSwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.err)), nil, unsafe.Pointer(&err))
}

// unblock call a service which has a reply
// if method, argv and reply do not match, return return a error
func (c *Context) Go(method string, argv interface{}, reply interface{}, done chan *Call) (*Call, error) {
	dptor := c.owner.getDescriptor(method)
	if dptor == nil {
		return nil, fmt.Errorf("call unknown method:%s", method)
	}

	if !dptor.HasReply() {
		return nil, fmt.Errorf("canot call method %s, use invoke instead", method)
	}

	if !dptor.MatchArgType(reflect.TypeOf(argv)) || !dptor.MatchReplyType(reflect.TypeOf(reply)) {
		return nil, fmt.Errorf("call method %s with unmatch arg or reply", method)
	}

	var pack proto_base.Pack
	session := c.nextSession()
	pack.Session = proto.Int32(session)
	pack.Type = proto.Int32(dptor.Id)
	pack.Data, _ = proto.Marshal(argv.(proto.Message))

	if done == nil {
		done = make(chan *Call, 1)
	} else {
		if cap(done) == 0 {
			return nil, fmt.Errorf("call %s: done channel is unbuffered", method)
		}
	}

	call := &Call{
		Dptor: dptor,
		Argv:  argv,
		Reply: reply,
		Done:  done,
	}
	c.setSession(session, call)
	c.writePack(&pack)
	return call, nil
}

func (c *Context) MustGo(method string, argv interface{}, reply interface{}, done chan *Call) *Call {
	call, err := c.Go(method, argv, reply, done)
	if err != nil {
		log.Panic("MustGo failed: method:%s, argv:%v", method, argv)
	}
	return call
}

// block call a service which has a reply
// if method, argv and reply do not match, return return a error
func (c *Context) Call(method string, argv interface{}, reply interface{}) (*CallError, error) {
	call, err := c.Go(method, argv, reply, nil)
	if err != nil {
		return nil, err
	}
	call = <-call.Done
	return call.Error, nil
}

// same as Call except it panics if method, argv and reply do not match
func (c *Context) MustCall(method string, argv interface{}, reply interface{}) *CallError {
	callError, err := c.Call(method, argv, reply)
	if err != nil {
		log.Panic("MustCall failed: method:%s, argv:%v", method, argv)
	}
	return callError
}

// invoke a service which has not a reply
func (c *Context) Invoke(method string, argv interface{}) error {
	dptor := c.owner.getDescriptor(method)
	if dptor == nil {
		return fmt.Errorf("invoke unknown method:%s", method)
	}

	if dptor.HasReply() {
		return fmt.Errorf("canot invoke method %s, use call instead", method)
	}

	if !dptor.MatchArgType(reflect.TypeOf(argv)) {
		return fmt.Errorf("invoke method %s with unmatch argv", method)
	}

	var pack proto_base.Pack
	pack.Session = proto.Int32(0)
	pack.Type = proto.Int32(dptor.Id)
	pack.Data, _ = proto.Marshal(argv.(proto.Message))

	c.writePack(&pack)
	return nil
}

// same as invoke except it panics if any error happen
func (c *Context) MustInvoke(method string, argv interface{}) {
	err := c.Invoke(method, argv)
	if err != nil {
		log.Panic("MustInvoke failed: method:%s, argv:%v", method, argv)
	}
}

func (c *Context) writePack(pack *proto_base.Pack) {
	log.Printf("write pack:%d %d %d", pack.GetSession(), pack.GetType(), len(pack.GetData()))
	err := c.codec.WritePack(pack)
	if err != nil {
		c.setError(err)
		c.Close()
	}
}

func (c *Context) dispatchResponse(pack *proto_base.Pack) bool {
	log.Printf("dispatch response:%d", pack.GetSession())
	call := c.grabSession(pack.GetSession())
	if call == nil {
		return c.owner.onUnknownPack(c, pack)
	}

	packError := pack.GetError()
	if packError.GetFailed() {
		call.Error = NewRpcCallError(packError.GetCode(), packError.GetError())
	} else {
		if err := proto.Unmarshal(pack.GetData(), call.Reply.(proto.Message)); err != nil {
			call.Error = NewCallError(0, err.Error())
		}
	}

	log.Printf("response done:%s, %v", call.Dptor.NormalName, call.Error)
	call.done()
	return true
}

func (c *Context) dispatchRequest(pack *proto_base.Pack) bool {
	log.Printf("dispatch request:%d %d %d", pack.GetSession(), pack.GetType(), len(pack.GetData()))
	typ := pack.GetType()
	s := c.owner.getService(typ)
	if s == nil {
		return c.owner.onUnknownPack(c, pack)
	}

	argv := reflect.New(s.dptor.ArgType.Elem())
	if err := proto.Unmarshal(pack.GetData(), argv.Interface().(proto.Message)); err != nil {
		return c.owner.onUnknownPack(c, pack)
	}

	var replyv reflect.Value
	if s.hasReply() {
		replyv = reflect.New(s.dptor.ReplyType.Elem())
	}

	go func() {
		if s.hasReply() {
			callError := s.call(c, argv, replyv)
			pack.Type = proto.Int32(0)
			pack.Data = nil
			if callError != nil {
				pack.GetError().Failed = proto.Bool(false)
				pack.GetError().Code = proto.Int32(callError.Code)
				pack.GetError().Error = proto.String(callError.Msg)
			} else {
				pack.Data, _ = proto.Marshal(replyv.Interface().(proto.Message))
			}
			c.writePack(pack)
		} else {
			s.invoke(c, argv)
		}
	}()
	log.Printf("request done:%d %d %d", pack.GetSession(), pack.GetType(), len(pack.GetData()))
	return true
}

func (c *Context) Close() error {
	c.closeAllSessions()
	return c.codec.Close()
}

func (c *Context) serve() {
	var err error
	for {
		var pack proto_base.Pack
		if err = c.codec.ReadPack(&pack); err != nil {
			c.owner.onIoError(c, err)
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

	c.setError(err)
	c.Close()
}
