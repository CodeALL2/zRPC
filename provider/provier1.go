package main

import (
	"fmt"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
	imp2 "zRPC/provider/imp"
	"zRPC/server/imp"
)

type TwoHandler struct {
	znet.BaseRouter
}

func (this *TwoHandler) PreHandler(request ziface.IRequest) {
	fmt.Println("调用之前的函数", request.GetConnection().GetRemoteAddr().String())
}

func (this *TwoHandler) Handler(request ziface.IRequest) {
	fmt.Println("方法2的handler执行", request.GetConnection().GetRemoteAddr().String())
	fmt.Println("数据体:", string(request.GetData()))
	err := request.GetConnection().SendMessage(request.GetMsgID(), []byte("hello this is zinx"))
	if err != nil {
		fmt.Println(err)
	}
	//开始拆
}

func (this *TwoHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
}

func main() {
	registry := imp.NewRegistry()

	registry.LocalRegistry("IUserService", &imp2.UserService{Id: 1, Name: "chu", Email: "xxx"})
	server := imp.NewServer()
	server.SetRegistry(registry)
	server.Start()
}
