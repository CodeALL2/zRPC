package imp

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"zRPC/protobufdemo/pp/pb"
)

type UserService struct {
	Id    int64
	Name  string
	Email string
}

func (us *UserService) GetUser(ctx context.Context, req *pb.RpcUser) (*pb.RpcResponse, error) {
	fmt.Println("req name", req.Name)
	rpcUser := &pb.RpcUser{
		Id:    us.Id,
		Name:  us.Name,
		Email: us.Email,
		Age:   12,
	}

	anyUser, err := anypb.New(rpcUser)

	if err != nil {
		return nil, err
	}
	response := pb.RpcResponse{
		TypeName:        "RpcUser",
		ResponseValue:   anyUser,
		ResponseMessage: "调用成功",
		ResponseStatue:  200,
	}
	return &response, nil
}
