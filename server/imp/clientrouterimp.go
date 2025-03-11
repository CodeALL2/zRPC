package imp

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/iface"
	"zRPC/server/util"
)

type ClientRPCRouterImp struct {
}

func (c *ClientRPCRouterImp) Handler(msg iface.IClientMsgRequest) {
	body := msg.GetData().GetMsg() //得到返回的数据
	//将这个投入到工作池进行处理消息
	//body就是服务端返回的信息
	response := &pb.RpcResponse{}
	if err := proto.Unmarshal(body, response); err != nil {
		fmt.Println("protobuf解析服务器返回的消息失败")
		c.Close(msg, response.MsgId) //关闭资源
		return
	}

	fmt.Println("返回的类型名:", response.TypeName)

	result, err := util.NewTypeRegistry().GetType(response.TypeName)
	if err != nil {
		fmt.Println("传回来的协议没有遵循protobuf")
		c.Close(msg, response.MsgId) //关闭资源
		return
	}

	if err = response.ResponseValue.UnmarshalTo(result); err != nil {
		fmt.Println("客户端解析消息protobuf返回结果数据失败")
		c.Close(msg, response.MsgId)
		return
	}
	//将值赋给结果集
	msg.GetMessageMap()[response.MsgId].Result <- result
}

func (c *ClientRPCRouterImp) Close(msg iface.IClientMsgRequest, msgId string) {
	close(msg.GetMessageMap()[msgId].Result) //关闭这个返回通道
	msg.GetMessageMap()[msgId] = nil
}
