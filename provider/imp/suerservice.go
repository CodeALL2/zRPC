package imp

import (
	"context"
	"fmt"
	"zRPC/protobufdemo/pp/pb"
)

type UserService struct {
	Id    int64
	Name  string
	Email string
}

func (us *UserService) GetUser(ctx context.Context, req *pb.VoidValue) (*pb.RpcUser, error) {
	rpcUser := &pb.RpcUser{
		Id:    us.Id,
		Name:  us.Name,
		Email: us.Email,
		Age:   12,
	}
	fmt.Println("方法已经被成功调用", rpcUser)

	return rpcUser, nil
}
