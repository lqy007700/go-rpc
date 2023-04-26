package rpc_i

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_setFuncField(t *testing.T) {
	tests := []struct {
		name    string
		server  Service
		wantErr bool
	}{
		{
			name:    "123",
			server:  &UserService{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(":8082")
			if err := setFuncField(tt.server, c); (err != nil) != tt.wantErr {
				t.Errorf("setFuncField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSend(t *testing.T) {
	// 服务端
	srv := NewServer()
	srv.registerServer(&UserServiceLocal{})
	go srv.Start(":8082")
	time.Sleep(time.Second * 2)

	// 客户端
	c := &UserService{}
	InitClientProxy(":8082", c)
	id, err := c.GetById(context.Background(), &GetByIdReq{})
	if err != nil {
		return
	}
	fmt.Println(id)
}
