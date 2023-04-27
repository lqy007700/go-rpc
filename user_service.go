package go_rpc

import (
	"context"
	"log"
)

type UserService struct {
	// 用反射来赋值
	// 本质上是一个字段
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}
func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(111, req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}
