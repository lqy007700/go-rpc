package go_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"go-rpc/message"
	"net"
	"reflect"
	"time"
)

const (
	numOfLengthBytes = 8
)

func InitClientProxy(addr string, service Service) error {
	client := NewClient(addr)
	return setFuncField(service, client)
}

// 设置请求server的func
func setFuncField(service Service, p Proxy) error {
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

				marshal, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldType.Name,
					Data:        marshal,
				}

				// 发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				err = json.Unmarshal(resp.Data, out)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{reflect.ValueOf(out), reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}

			fnVal := reflect.MakeFunc(fieldType.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

type Client struct {
	addr string
	t    time.Duration
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		t:    time.Second * 3,
	}
}

// Invoke 发送请求到服务端
func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.send(data)
	if err != nil {
		return nil, err
	}

	return &message.Response{
		Data: resp,
	}, nil
}

func (c *Client) send(data []byte) ([]byte, error) {
	// 建立链接
	conn, err := net.DialTimeout("tcp", c.addr, c.t)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	req, err := EncodeMsg(data)
	if err != nil {
		return nil, err
	}

	// 发送请求
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	resp, err := ReadMsg(conn)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
