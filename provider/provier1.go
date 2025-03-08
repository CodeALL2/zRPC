package main

import (
	"fmt"
	imp2 "zRPC/provider/imp"
	"zRPC/server/imp"
	"zRPC/server/model"
	"zRPC/server/util"
)

func main() {
	zRCApplication := imp.NewZRPCApplication()
	registry := imp.NewRegistry()
	//注册本地服务
	registry.LocalRegistry("IUserService", &imp2.UserService{Id: 1, Name: "chu", Email: "xxx"})
	//注册中心
	registryConfig := zRCApplication.GetRegistryConfig()
	registryConfig.SetRegistry("etcd-server")
	registryConfig.SetRegistryAddr("localhost:2379")
	registryConfig.SetTimeOut(30)
	registryFactory := zRCApplication.GetRegistryFactory()
	registryServer := registryFactory.GetRegistryServer(registryConfig.GetRegistryName())
	err := registryServer.Init(registryConfig)

	if err != nil {
		fmt.Println("连接注册中心错误")
		return
	}
	info := &model.ServiceMetaInfo{
		ServiceName:    "IUserService",
		ServiceHost:    "localhost",
		ServicePort:    "9888",
		ServiceVersion: "v1.0",
	}

	err = registryServer.Register(info)
	if err != nil {
		fmt.Println("注册元数据失败")
		return
	}

	util.ConfigPath = "F:\\zRPC\\provider\\conf\\tsconfig.json"
	server := imp.NewServer()
	server.SetRegistry(registry)
	server.SetRegistryServer(registryServer)
	server.Start()
}
