package message

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
