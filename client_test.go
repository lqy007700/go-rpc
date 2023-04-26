package go_rpc

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	server.RegisterService(&UserServiceServer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	err := InitClientProxy(":8081", usClient)
	if err != nil {
		return
	}

	resp, err := usClient.GetById(context.Background(), &GetByIdReq{
		Id: 123,
	})
	log.Println(resp)
	if err != nil {
		return
	}
}
