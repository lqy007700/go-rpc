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
		err := server.Start("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	c := NewClient(":8082")
	usClient := &UserService{}
	err := c.InitService(usClient)
	if err != nil {
		t.Log(err)
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
