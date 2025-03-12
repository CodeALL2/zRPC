package imp

import (
	"fmt"
	"time"
	"zRPC/server/iface"
)

type ClientHeartRouterImp struct {
}

func (c *ClientHeartRouterImp) Handler(msg iface.IClientMsgRequest) {
	//心跳包的handler
	data := msg.GetData().GetMsg()
	key := string(data)
	msg.GetHeartServerMap().Store(key, time.Now().Unix())
	fmt.Println("节点", key, "已回复心跳")
}
