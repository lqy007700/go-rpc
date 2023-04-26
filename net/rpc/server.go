package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service, 8),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
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
		// 请求内容长度
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		length := binary.BigEndian.Uint64(lenBs)

		// 请求内容
		resBs := make([]byte, length)
		_, err = conn.Read(resBs)
		if err != nil {
			return err
		}

		// 处理请求
		respData, err := s.handlerMsg(resBs)
		if err != nil {
			return err
		}

		// 发送响应
		respLen := len(respData)
		res := make([]byte, respLen+numOfLengthBytes)
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		copy(res[numOfLengthBytes:], respData)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

// 处理请求信息，还原调用信息，发起本地业务调用
func (s *Server) handlerMsg(data []byte) ([]byte, error) {
	req := &Request{}
	err := json.Unmarshal(data, req)
	if err != nil {
		return nil, err
	}

	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("服务不存在")
	}

	// 反射找到方法，并且执行调用
	val := reflect.ValueOf(service)
	method := val.MethodByName(req.MethodName)

	// 请求参数
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())

	err = json.Unmarshal(req.Args, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq

	result := method.Call(in)

	if result[1].Interface() != nil {
		return nil, result[1].Interface().(error)
	}

	resp, err := json.Marshal(result[0].Interface())
	if err != nil {
		return nil, err
	}

	return resp, nil
}
