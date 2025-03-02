package imp

import (
	"context"
	"zRPC/protobufdemo/pp/pb"
)

type UserService struct {
	Id    int64
	Name  string
	Email string
}

func (us *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return &pb.GetUserResponse{
		User: &pb.User{
			Id:    us.Id,
			Name:  us.Name,
			Email: us.Email,
		},
		Success: true,
	}, nil
}
