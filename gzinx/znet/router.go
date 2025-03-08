package znet

import (
	"fmt"
	"zRPC/gzinx/ziface"
)

type BaseRouter struct{}

func (router *BaseRouter) PreHandler(request ziface.IRequest) {
	fmt.Printf("默认方法")
}
func (router *BaseRouter) Handler(request ziface.IRequest) {
	fmt.Printf("默认方法")
}
func (router *BaseRouter) PostHandler(request ziface.IRequest) {
	fmt.Printf("默认方法")
}
