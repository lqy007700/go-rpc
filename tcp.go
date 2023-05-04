package go_rpc

import (
	"encoding/binary"
	"log"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 响应内容长度
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		log.Println(6, err.Error())
		return nil, err
	}
	headLen := binary.BigEndian.Uint32(lenBs[:4])
	bodyLen := binary.BigEndian.Uint32(lenBs[4:])
	length := headLen + bodyLen

	// 响应内容
	resp := make([]byte, length)
	_, err = conn.Read(resp[8:])
	copy(resp[:8], lenBs)
	return resp, err
}
