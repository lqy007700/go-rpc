package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeData(t *testing.T) {
	tests := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				HeadLen:     0,
				BodyLen:     100,
				RequestId:   1,
				Version:     1,
				Compress:    1,
				Serialize:   32,
				ServiceName: "my-service",
				MethodName:  "my-name",
				Mate:        map[string]string{"name": "liu"},
				Data:        []byte("liu"),
			},
		},
		{
			name: "normal 1",
			req: &Request{
				HeadLen:     0,
				BodyLen:     100,
				RequestId:   1,
				Version:     1,
				Compress:    1,
				Serialize:   32,
				ServiceName: "my-service",
				MethodName:  "my-name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.calculateHeadLen()
			tt.req.calculateBodyLen()

			req := EncodeReq(tt.req)
			decodeReq := DecodeReq(req)
			assert.Equal(t, tt.req, decodeReq)
		})
	}
}
