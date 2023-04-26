package rpc_i

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
		services: make(map[string]Service, 10),
	}
}

func (s *Server) registerServer(server Service) {
	s.services[server.Name()] = server
}

func (s *Server) Start(addr string) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			return
		}

		go func() {
			if err := s.handlerConn(conn); err != nil {
				conn.Close()
			}
		}()
	}
}

func (s *Server) handlerConn(conn net.Conn) error {
	for {
		readLen := make([]byte, sendNumLen)
		_, err := conn.Read(readLen)
		if err != nil {
			return err
		}

		// 获取请求信息
		length := binary.BigEndian.Uint64(readLen)
		readData := make([]byte, length)
		_, err = conn.Read(readData)
		if err != nil {
			return err
		}

		respData, err := s.handlerMsg(readData)
		if err != nil {
			return err
		}

		respLen := len(respData)
		resp := make([]byte, respLen+sendNumLen)
		binary.BigEndian.PutUint64(resp[:sendNumLen], uint64(respLen))
		copy(resp[sendNumLen:], respData)
		_, err = conn.Write(resp)
		if err != nil {
			return err
		}

	}
}

// 处理请求信息 还原调用信息，发起本地调用 获取响应
func (s *Server) handlerMsg(data []byte) ([]byte, error) {
	req := &Request{}
	err := json.Unmarshal(data, req)
	if err != nil {
		return nil, err
	}

	// 获取本地服务
	server, ok := s.services[req.ServerName]
	if !ok {
		return nil, errors.New("服务不存在")
	}

	// 反射找到本地方法
	of := reflect.ValueOf(server)
	method := of.MethodByName(req.MethodName)

	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())

	inReq := reflect.New(method.Type().In(1).Elem())

	err = json.Unmarshal(req.Args, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq

	// 发起本地调用
	// 结果，err
	result := method.Call(in)
	if result[1].Interface() != nil {
		return nil, result[1].Interface().(error)
	}

	marshal, err := json.Marshal(result[0].Interface())
	if err != nil {
		return nil, err
	}
	return marshal, nil
}
