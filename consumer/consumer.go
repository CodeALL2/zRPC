package main

import (
	"fmt"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/imp"
)

func main() {
	zRPC := imp.NewRPCClient("127.0.0.1", "9888")
	if err := zRPC.Dial(); err != nil {
		fmt.Println("客户端连接失败", err)
		return
	}
	result, err := zRPC.Invoke("IUserService", "GetUser", "VoidValue", &pb.VoidValue{})
	if err != nil {
		fmt.Println(err)
	}
	user := result.(*pb.RpcUser)
	fmt.Println("返回值", user)
}
