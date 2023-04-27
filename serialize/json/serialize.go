package json

import "encoding/json"

type Serialize struct {
}

func (s *Serialize) Code() uint8 {
	return 1
}

func (s *Serialize) Encode(val any) ([]byte, error) {
	return json.Marshal(val)
}

func (s *Serialize) Decode(data []byte, val any) error {
	return json.Unmarshal(data, val)
}
