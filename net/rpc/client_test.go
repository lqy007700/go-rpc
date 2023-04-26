package rpc

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestSetFuncField(t *testing.T) {
	//tests := []struct {
	//	name string
	//
	//	mock func(ctrl *gomock.Controller) Proxy
	//
	//	service Service
	//	wantErr error
	//}{
	//	{
	//		name: "userservice",
	//		mock: func(ctrl *gomock.Controller) Proxy {
	//			p := NewMockProxy(ctrl)
	//			p.EXPECT().Invoke(gomock.Any(), &Request{
	//				ServiceName: "user-service",
	//				MethodName:  "GetById",
	//				Args: &GetByIdReq{
	//					Id: 123,
	//				},
	//			}).Return(&Response{}, nil)
	//			return p
	//		},
	//		service: &UserService{},
	//		wantErr: nil,
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		ctrl := gomock.NewController(t)
	//		defer ctrl.Finish()
	//
	//		if err := setFuncField(tt.service, tt.mock(ctrl)); err != nil {
	//			t.Errorf("InitClientProxy() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//
	//		id, err := tt.service.(*UserService).GetById(context.Background(), &GetByIdReq{Id: 123})
	//		if err != nil {
	//			panic(err)
	//			return
	//		}
	//		t.Log(id)
	//	})
	//}
}

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
