package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
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
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldType.Name,
					Args:        marshal,
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
func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.send(data)
	if err != nil {
		return nil, err
	}

	return &Response{
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

	// 计算请求数据长度
	reqLen := len(data)

	// 请求数据 长度 + 内容
	req := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	copy(req[numOfLengthBytes:], data)

	// 发送请求
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	// 响应内容长度
	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint64(lenBs)

	// 响应内容
	resp := make([]byte, length)
	_, err = conn.Read(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
