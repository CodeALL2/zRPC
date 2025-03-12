package iface

import (
	"net"
	"sync"
	"zRPC/gzinx/ziface"
	"zRPC/server/model"
)

type IClientMsgRequest interface {
	GetConnection() net.Conn
	GetData() ziface.IMessage
	GetMessageMap() map[string]*model.MsgResult
	GetHeartServerMap() *sync.Map
}
