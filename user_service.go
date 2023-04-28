package go_rpc

import (
	"context"
	"go-rpc/proto/gen"
	"log"
)

type UserService struct {
	// 用反射来赋值
	// 本质上是一个字段
	GetById      func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
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
	log.Println("GetById-local", req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println("GetByIdProto--Local", req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   1,
			Name: "liu",
		},
	}, u.Err
}
