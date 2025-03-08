package znet

import "zRPC/gzinx/ziface"

type Request struct {
	conn ziface.IConnection
	data ziface.IMessage
}

func (req *Request) GetConnection() ziface.IConnection {
	return req.conn
}

func (req *Request) GetData() []byte {
	return req.data.GetMsg()
}

func (req *Request) GetMsgID() uint32 {
	return req.data.GetMsgId()
}

func (req *Request) GetMsgLen() uint32 {
	return req.data.GetMsgLen()
}
