package model

import (
	"net"
	"zRPC/gzinx/ziface"
)

type ClientMsgRequest struct {
	MessageMap map[string]*MsgResult //存放服务的所有消息
	Con        net.Conn              //连接对象
	Data       ziface.IMessage       //真正的消息体
}

func (c *ClientMsgRequest) GetConnection() net.Conn {
	return c.Con
}
func (c *ClientMsgRequest) GetData() ziface.IMessage {
	return c.Data
}

func (c *ClientMsgRequest) GetMessageMap() map[string]*MsgResult {
	return c.MessageMap
}
