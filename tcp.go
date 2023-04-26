package go_rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 响应内容长度
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint64(lenBs)

	// 响应内容
	resp := make([]byte, length)
	_, err = conn.Read(resp)
	return resp, err
}

func EncodeMsg(data []byte) ([]byte, error) {
	// 计算请求数据长度
	reqLen := len(data)

	// 请求数据 长度 + 内容
	res := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(reqLen))
	copy(res[numOfLengthBytes:], data)
	return res, nil
}
