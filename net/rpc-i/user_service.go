package rpc_i

import (
	"context"
	"fmt"
)

// UserService 远程服务对应的本地服务
// 需要获取到请求到服务和请求到方法
type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u *UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
}

type GetByIdResp struct {
	Id int
}

type UserServiceLocal struct {
}

func (u *UserServiceLocal) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	fmt.Println("get by id")
	return &GetByIdResp{
		Id: 123123,
	}, nil
}

func (u *UserServiceLocal) Name() string {
	return "user-service"
}
