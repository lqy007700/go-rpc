package rpc_i

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

const (
	sendNumLen = 8
)

func InitClientProxy(addr string, server Service) {
	client := NewClient(addr)
	err := setFuncField(server, client)
	if err != nil {
		return
	}
}

// 设置请求server的func
func setFuncField(server Service, p Proxy) error {
	val := reflect.ValueOf(server)

	typ := val.Type()
	// 只允许指向结构体的指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("只允许指向结构体的指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		typField := typ.Field(i)
		valField := val.Field(i)

		if valField.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {

				// 获取成员方法第一个返回值类型 ex GetByIdResp{}
				returnArg := typField.Type.Out(0)
				// 作空返回值用
				retVal := reflect.New(returnArg).Elem()
				// unmarshal 响应用
				out := reflect.New(returnArg.Elem()).Interface()
				ctx := args[0].Interface().(context.Context)

				// 请求参数
				reqArg, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				req := &Request{
					ServerName: server.Name(),
					MethodName: typField.Name,
					Args:       reqArg,
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

			makeFunc := reflect.MakeFunc(typField.Type, fn)
			valField.Set(makeFunc)
		}
	}
	return nil
}

type Client struct {
	addr string
}

// Invoke 发送请求
func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.send(reqData)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, nil
}

func (c *Client) send(data []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, nil
	}
	defer conn.Close()

	// 发送数据
	reqLen := len(data)
	reqData := make([]byte, reqLen+sendNumLen)

	binary.BigEndian.PutUint64(reqData[:sendNumLen], uint64(reqLen))
	copy(reqData[sendNumLen:], data)

	_, err = conn.Write(reqData)
	if err != nil {
		return nil, nil
	}

	respLen := make([]byte, sendNumLen)
	_, err = conn.Read(respLen)
	if err != nil {
		return nil, nil
	}

	length := binary.BigEndian.Uint64(respLen)
	respData := make([]byte, length)
	_, err = conn.Read(respData)
	if err != nil {
		return nil, nil
	}

	return respData, nil
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}
