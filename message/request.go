package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	// 请求头
	HeadLen     uint32            // 请求头长度
	BodyLen     uint32            // 请求体长度
	RequestId   uint32            // 请求ID
	Version     uint8             // 版本
	Compress    uint8             // 压缩算法
	Serialize   uint8             // 序列化算法
	ServiceName string            // 服务名
	MethodName  string            // 方法名
	Mate        map[string]string // 自定义元数据

	// 请求体
	Data []byte // 参数
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLen+req.BodyLen)

	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLen)
	// 2. 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLen)
	// 3. req id
	binary.BigEndian.PutUint32(bs[8:12], req.RequestId)
	// 4. 版本
	bs[12] = req.Version
	// 5. 压缩算法
	bs[13] = req.Compress
	// 6. 序列化协议
	bs[14] = req.Serialize

	// 7. 服务名称
	cur := bs[15:]
	copy(cur, req.ServiceName)
	// 分割符
	cur = cur[len(req.ServiceName):]
	cur[0] = '\n'
	cur = cur[1:]
	// 8. 方法名称
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	// 9. mate
	cur[0] = '\n'
	cur = cur[1:]
	for k, v := range req.Mate {
		copy(cur, k)
		cur = cur[len(k):]
		cur[0] = ':'
		cur = cur[1:]
		copy(cur, v)
		cur = cur[len(v):]
		cur[0] = '\n'
		cur = cur[1:]
	}

	copy(cur, req.Data)
	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1. 头部长度
	req.HeadLen = binary.BigEndian.Uint32(data[:4])
	// 2. body长度
	req.BodyLen = binary.BigEndian.Uint32(data[4:8])
	// 3. req id
	req.RequestId = binary.BigEndian.Uint32(data[8:12])
	// 4. 版本
	req.Version = data[12]
	// 5. 压缩算法
	req.Compress = data[13]
	// 6. 序列化协议
	req.Serialize = data[14]

	header := data[15:req.HeadLen]
	// 7. 分割服务名与方法名user-service:userMethod
	idx := bytes.IndexByte(header, '\n')
	// 8.服务名 方法名
	req.ServiceName = string(header[:idx])
	header = header[idx+1:]

	idx = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:idx])
	header = header[idx+1:]

	// 9.meta
	idx = bytes.IndexByte(header, '\n')
	if idx != -1 {
		meta := make(map[string]string, 4)

		for idx != -1 {
			pair := header[:idx]
			pairIdx := bytes.IndexByte(pair, ':')
			meta[string(pair[:pairIdx])] = string(pair[pairIdx+1:])

			header = header[idx+1:]
			idx = bytes.IndexByte(header, '\n')
		}
		req.Mate = meta
	}

	if req.BodyLen != 0 {
		req.Data = data[req.HeadLen:]
	}
	return req
}

func (r *Request) CalculateHeadLen() {
	headLen := 15 + len(r.ServiceName) + 1 + len(r.MethodName) + 1
	for k, v := range r.Mate {
		headLen += len(k) + 1 + len(v) + 1
	}
	r.HeadLen = uint32(headLen)
}

func (r *Request) CalculateBodyLen() {
	r.BodyLen = uint32(len(r.Data))
}
