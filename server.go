package go_rpc

import (
	"context"
	"errors"
	"go-rpc/message"
	"go-rpc/serialize"
	"go-rpc/serialize/json"
	"net"
	"reflect"
)

type Server struct {
	services   map[string]*reflectionStub
	serializes map[uint8]serialize.Serialize
}

func NewServer() *Server {
	res := &Server{
		services:   make(map[string]*reflectionStub, 8),
		serializes: make(map[uint8]serialize.Serialize, 4),
	}
	res.RegisterSerialize(&json.Serialize{})
	return res
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = &reflectionStub{
		s:          service,
		value:      reflect.ValueOf(service),
		serializes: s.serializes,
	}
}

func (s *Server) RegisterSerialize(sl serialize.Serialize) {
	s.serializes[sl.Code()] = sl
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

	respData, err := service.invoke(ctx, req)
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
	s          Service
	value      reflect.Value
	serializes map[uint8]serialize.Serialize
}

func (r *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := r.value.MethodByName(req.MethodName)

	// 请求参数
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())

	s, ok := r.serializes[req.Serialize]
	if !ok {
		return nil, errors.New("序列化协议不存在")
	}

	err := s.Decode(req.Data, inReq.Interface())
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
		res, er = s.Encode(result[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, nil
}
