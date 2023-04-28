package proto

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

type Serialize struct {
}

func (s *Serialize) Code() uint8 {
	return 2
}

func (s *Serialize) Encode(val any) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("必须是proto.Message")
	}
	return proto.Marshal(msg)
}

func (s *Serialize) Decode(data []byte, val any) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("必须是proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
