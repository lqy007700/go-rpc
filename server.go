package go_rpc

import (
	"context"
	"encoding/json"
	"errors"
	"go-rpc/message"
	"net"
	"reflect"
)

type Server struct {
	services map[string]*reflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*reflectionStub, 8),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = &reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
}

func (s *Server) Start(network, addr string) error {
	listen, err := net.Listen(network, addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}

		go func() {
			err := s.handlerConn(conn)
			if err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handlerConn(conn net.Conn) error {
	for {
		resBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		req := message.DecodeReq(resBs)
		respData, err := s.Invoke(context.Background(), req)
		if err != nil {
			respData.Error = []byte(err.Error())
		}

		respData.CalculateHeadLen()
		respData.CalculateBodyLen()
		_, err = conn.Write(message.EncodeResp(respData))
		if err != nil {
			return err
		}
	}
}

// Invoke 处理请求信息，还原调用信息，发起本地业务调用
func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("服务不存在")
	}

	respData, err := service.invoke(ctx, req.MethodName, req.Data)
	if err != nil {
		return nil, err
	}

	resp := &message.Response{
		RequestId: req.RequestId,
		Version:   req.Version,
		Compress:  req.Compress,
		Serialize: req.Serialize,
		Data:      respData,
		Error:     nil,
	}
	return resp, nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (r *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := r.value.MethodByName(methodName)

	// 请求参数
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())

	err := json.Unmarshal(data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq

	result := method.Call(in)

	if result[1].Interface() != nil {
		err = result[1].Interface().(error)
	}

	var res []byte
	if result[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = json.Marshal(result[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, nil
}
