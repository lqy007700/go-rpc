package serialize

type Serialize interface {
	Code() uint8
	Encode(val any) ([]byte, error)
	// Decode val结构体指针
	Decode(data []byte, val any) error
}
