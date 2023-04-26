package rpc

import "context"

type Service interface {
	Name() string
}

type Request struct {
	ServiceName string
	MethodName  string
	Args        []byte
}

type Response struct {
	Data []byte
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}
