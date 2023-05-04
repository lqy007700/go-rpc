package go_rpc

import (
	"context"
	"errors"
	"go-rpc/message"
	"go-rpc/serialize"
	"go-rpc/serialize/json"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

const (
	numOfLengthBytes = 8
)

type ClientOpt func(c *Client)

func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serialize)
}

// 设置请求server的func
func setFuncField(service Service, p Proxy, s serialize.Serialize) error {
	val := reflect.ValueOf(service)

	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("只支持指向结构体的一级指针")
	}

	val = val.Elem() // nil
	typ = typ.Elem() // client.UserService

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldType := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {
				retVal := reflect.New(fieldType.Type.Out(0)).Elem()
				out := reflect.New(fieldType.Type.Out(0).Elem()).Interface()

				ctx := args[0].Interface().(context.Context)

				marshal, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				meta := make(map[string]string, 2)
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.UnixMilli(), 10)
				}

				if isOneway(ctx) {
					meta["oneway"] = "ok"
				}
				req := &message.Request{
					Serialize:   s.Code(),
					ServiceName: service.Name(),
					MethodName:  fieldType.Name,
					Data:        marshal,
					Mate:        meta,
				}
				req.CalculateHeadLen()
				req.CalculateBodyLen()

				// 发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				// oneway
				if resp == nil {
					return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
				}

				var serverErr error
				if len(resp.Error) > 0 {
					serverErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					err = s.Decode(resp.Data, out)
					if err != nil {
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if serverErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(serverErr)
				}

				return []reflect.Value{reflect.ValueOf(out), retErrVal}
			}

			fnVal := reflect.MakeFunc(fieldType.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

type Client struct {
	addr      string
	t         time.Duration
	serialize serialize.Serialize
}

func ClientWithSerialize(sl serialize.Serialize) ClientOpt {
	return func(c *Client) {
		c.serialize = sl
	}
}

func NewClient(addr string, opts ...ClientOpt) *Client {
	res := &Client{
		addr:      addr,
		t:         time.Second * 3,
		serialize: &json.Serialize{},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

// Invoke 发送请求到服务端
func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	// 检测超时
	//if ctx.Err() != nil {
	//	return nil, ctx.Err()
	//}

	var (
		resp *message.Response
		err  error
		ch   = make(chan struct{})
	)

	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return resp, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)

	conn, err := c.send(ctx, data)
	if isOneway(ctx) {
		log.Println("client oneway")
		return nil, nil
	}
	resp, err := ReadMsg(conn)
	if err != nil {
		log.Println(9, err.Error())
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) (net.Conn, error) {
	// 建立链接
	conn, err := net.DialTimeout("tcp", c.addr, c.t)
	if err != nil {
		return nil, err
	}

	// 发送请求
	_, err = conn.Write(data)
	if err != nil {
		log.Println(7, err.Error())
		return nil, err
	}
	return conn, nil
}
