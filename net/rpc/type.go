package rpc

import "context"

type Service interface {
	Name() string
}

type Request struct {
	HeadLen   uint32 // 请求头长度
	BodyLen   uint32 // 请求体长度
	RequestId uint32 // 请求ID
	Version   uint8  // 版本
	Compress  uint8  // 压缩算法
	Serialize uint8  // 序列化算法

	ServiceName string            // 服务名
	MethodName  string            // 方法名
	Mate        map[string]string // 自定义元数据

	Data []byte // 参数
}

type Response struct {
	HeadLen   uint32 // 请求头长度
	BodyLen   uint32 // 请求体长度
	RequestId uint32 // 请求ID
	Version   uint8  // 版本
	Compress  uint8  // 压缩算法
	Serialize uint8  // 序列化算法

	Data  []byte
	Error []byte
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}
