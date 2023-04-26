package rpc_i

import "context"

type Service interface {
	Name() string
}

type Request struct {
	ServerName string
	MethodName string
	Args       []byte
}

type Response struct {
	Data []byte
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}
