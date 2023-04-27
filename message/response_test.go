package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResData(t *testing.T) {
	tests := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				HeadLen:   0,
				BodyLen:   100,
				RequestId: 1,
				Version:   1,
				Compress:  1,
				Serialize: 32,
				Data:      []byte("liu"),
			},
		},
		{
			name: "err",
			resp: &Response{
				HeadLen:   0,
				BodyLen:   100,
				RequestId: 1,
				Version:   1,
				Compress:  1,
				Serialize: 32,
				Error:     []byte("错了"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.resp.CalculateHeadLen()
			tt.resp.CalculateBodyLen()

			req := EncodeResp(tt.resp)
			decodeReq := DecodeResp(req)
			assert.Equal(t, tt.resp, decodeReq)
		})
	}
}
