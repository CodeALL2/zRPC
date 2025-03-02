package ziface

import "net"

// 连接接口的封装
type IConnection interface {
	//连接启动
	Star()
	//连接关闭
	Stop()
	//获取连接套接字
	GetTCPConnection() net.Conn
	//获取连接id
	GetConnID() uint32
	//获取客户端的地址
	GetRemoteAddr() net.Addr
	//发送数据
	Send([]byte) error
	//将数据写到通道里
	SendMessage(msgId uint32, data []byte) error
	GetProperty(name string) interface{}              //获取属性
	SetProperty(name string, value interface{}) error //添加属性
	RemoveProperty(name string)
	GetServer() IServer
}
