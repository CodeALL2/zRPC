package imp

import (
	"fmt"
	"zRPC/gzinx/ziface"
	"zRPC/gzinx/znet"
	"zRPC/server/iface"
)

type HeartHandler struct {
	znet.BaseRouter
	Server iface.IServer
}

func (h *HeartHandler) Handler(request ziface.IRequest) {
	fmt.Println("HeartHandler Handle")
	fmt.Println("节点", string(request.GetData()), "已收到心跳包")
	request.GetConnection().SendMessage(2, request.GetData())
}
