package main

import (
	"fmt"
	"zRPC/gzinx/utils"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
)

type UserHandler struct {
	znet.BaseRouter
}

func (this *UserHandler) PreHandler(request ziface.IRequest) {
	request.GetData()
	fmt.Println("调用之前的函数", request.GetConnection().GetRemoteAddr().String())
}

func (this *UserHandler) Handler(request ziface.IRequest) {
	fmt.Println("方法1的handler执行", request.GetConnection().GetRemoteAddr().String())
	fmt.Println("数据体:", string(request.GetData()))
	//将数据回写给客户端
	err := request.GetConnection().SendMessage(request.GetMsgID(), []byte("hello this is zinx"))
	if err != nil {
		fmt.Println(err)
	}
}

func (this *UserHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
}

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
}

func (this *TwoHandler) PostHandler(request ziface.IRequest) {
	fmt.Println("handler结束后的函数", request.GetConnection().GetRemoteAddr().String())
}

// 连接开始之后的
func ConnectionStart(c ziface.IConnection) {
	fmt.Println("连接方法已经调用")
	if err := c.SendMessage(202, []byte("hello this is zinx")); err != nil {
		fmt.Println(err)
	}
}

// 连接 关闭
func ConnectionStop(c ziface.IConnection) {
	fmt.Println("连接关闭方法已经调用")
	if err := c.SendMessage(202, []byte("zinx stop")); err != nil {
		fmt.Println(err)
	}
}

func main() {
	utils.Reload("F:\\go-tcp-server\\gzinx\\demo\\conf\\gzinx.json")
	server := znet.NewServer("GZINX SERVER")
	server.AddOnConnStartHook(ConnectionStart)
	server.AddOnConnStopHook(ConnectionStop)
	handle := znet.NewMsgHandle()
	handle.AddMsgHandler(1, &UserHandler{})
	handle.AddMsgHandler(2, &TwoHandler{})

	server.AddMsgHandler(handle)
	server.Serve()
}
