package main

import (
	"fmt"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/imp"
)

func main() {
	//注册中心的注册
	zrpcApplication := imp.NewZRPCApplication()
	registryConfig := zrpcApplication.GetRegistryConfig()
	registryConfig.SetRegistry("etcd")
	registryConfig.SetRegistryAddr("localhost:2379")
	registryConfig.SetTimeOut(30)

	zRPC := imp.NewRPCClient(zrpcApplication) //将注册中心注入到RPCClient中去

	result, err := zRPC.Invoke("IUserService", "GetUser", "VoidValue", &pb.VoidValue{})
	if err != nil {
		fmt.Println(err)
		return
	}
	user := result.(*pb.RpcUser)
	fmt.Println("返回值", user)
}
