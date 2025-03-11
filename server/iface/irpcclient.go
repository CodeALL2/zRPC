package iface

import (
	"net"
	"zRPC/server/model"
)

type IRPCClient interface {
	Dial(addr string, port string) (error, net.Conn)
	Invoke(serviceName string, methodName string, paramName string, value interface{}) (*model.MsgResult, error)
}
