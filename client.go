package go_rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-rpc/codec"
	"io"
	"log"
	"net"
	"sync"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Err           error
	Done          chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

var ErrShutDown = errors.New("connection is shut down")

var _ io.Closer = (*Client)(nil)

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closing {
		return ErrShutDown
	}

	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.closing && !c.shutdown
}

// 将参数 call 添加到 client.pending 中，并更新 client.seq。
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutDown
	}

	call.Seq = c.seq
	c.pending[c.seq] = call
	c.seq++
	return c.seq, nil
}

//根据 seq，从 client.pending 中移除对应的 call，并返回。
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

//服务端或客户端发生错误时调用，将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call。
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.shutdown = true
	for _, call := range c.pending {
		call.Err = err
		call.done()
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		h := &codec.Header{}
		if err := c.cc.ReadHeader(h); err != nil {
			break
		}

		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Err = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Err = errors.New("reading body " + err.Error())
			}
			call.done()
		}

	}
	c.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	log.Printf("%+v", opt)
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

func (c *Client) send(call *Call) {
	// make sure that the client will send a complete request
	c.sending.Lock()
	defer c.sending.Unlock()

	// register this call.
	seq, err := c.registerCall(call)
	if err != nil {
		call.Err = err
		call.done()
		return
	}

	// prepare request header
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	// encode and send the request
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		// call may be nil, it usually means that Write partially failed,
		// client has received the response and handled
		if call != nil {
			call.Err = err
			call.done()
		}
	}
}

func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Err
}
