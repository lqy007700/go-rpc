package message

import (
	"encoding/binary"
)

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

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLen+resp.BodyLen)

	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLen)
	// 2. 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLen)
	// 3. req id
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestId)
	// 4. 版本
	bs[12] = resp.Version
	// 5. 压缩算法
	bs[13] = resp.Compress
	// 6. 序列化协议
	bs[14] = resp.Serialize
	cur := bs[15:]
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)
	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	// 1. 头部长度
	resp.HeadLen = binary.BigEndian.Uint32(data[:4])
	// 2. body长度
	resp.BodyLen = binary.BigEndian.Uint32(data[4:8])
	// 3. req id
	resp.RequestId = binary.BigEndian.Uint32(data[8:12])
	// 4. 版本
	resp.Version = data[12]
	// 5. 压缩算法
	resp.Compress = data[13]
	// 6. 序列化协议
	resp.Serialize = data[14]
	if resp.HeadLen > 15 {
		resp.Error = data[15:resp.HeadLen]
	}

	if resp.BodyLen != 0 {
		resp.Data = data[resp.HeadLen:]
	}
	return resp
}

func (r *Response) CalculateHeadLen() {
	r.HeadLen = 15 + +uint32(len(r.Error))
}

func (r *Response) CalculateBodyLen() {
	r.BodyLen = uint32(len(r.Data))
}
