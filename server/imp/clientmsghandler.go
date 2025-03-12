package imp

import (
	"fmt"
	"zRPC/server/iface"
)

type ClientMsgHandle struct {
	routerMap       map[uint32]iface.IClientRouter
	ClientTaskQueue []chan iface.IClientMsgRequest
}

var index uint32 = 0

func NewClientMsgHandle() *ClientMsgHandle {
	c := &ClientMsgHandle{
		routerMap:       make(map[uint32]iface.IClientRouter),
		ClientTaskQueue: make([]chan iface.IClientMsgRequest, 5),
	}
	c.ClientStartWorkerPoll() //开启工作池
	return c
}

func (c *ClientMsgHandle) ClientStartWorkerPoll() {
	for i := 0; i < 5; i++ {
		c.ClientTaskQueue[i] = make(chan iface.IClientMsgRequest, 10)
		go c.ClientStartWorker(i, c.ClientTaskQueue[i])
	}
}

func (c *ClientMsgHandle) ClientStartWorker(workId int, channel chan iface.IClientMsgRequest) {
	fmt.Println("客户端工作池", workId, "已开启")
	for {
		select {
		case msgReq := <-channel:
			c.DoMsgHandler(msgReq)
		}
	}

}

func (c *ClientMsgHandle) DoMsgHandler(request iface.IClientMsgRequest) {
	router, ok := c.routerMap[request.GetData().GetMsgId()]
	if !ok {
		fmt.Println("没有发现此id", request.GetData().GetMsgId(), "的路由")
		return
	}
	router.Handler(request)
}
func (c *ClientMsgHandle) AddMsgHandler(routerId uint32, router iface.IClientRouter) {
	if c.routerMap == nil {
		fmt.Println("路由表没有初始化")
		return
	}
	c.routerMap[routerId] = router
}
func (c *ClientMsgHandle) SendMessageToClientQueue(request iface.IClientMsgRequest) {
	queue := c.ClientTaskQueue[index%5]
	index++
	queue <- request
}
