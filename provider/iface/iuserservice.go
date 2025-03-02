package iface

import (
	"context"
	"zRPC/protobufdemo/pp/pb"
)

type IUserService interface {
	GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error)
}
