package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net"
	"time"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/util"
)

type Proxy struct {
}

func (p *Proxy) proxy(param interface{}) {
}

type UserService struct {
	//集成一个类，这个类里面有一个实现的模板方法
	Proxy
}

func (this *UserService) GetUser(user *pb.RpcUser) {

	//方法名
	//结构体名
	//user *pb.RpcUser 类型名

	this.proxy(user)
}

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:9888")
	if err != nil {
		fmt.Println("连接失败", err)
	}
	defer conn.Close()
	user := &pb.RpcUser{Id: 2, Name: "xx", Email: "xxx"}
	anyUser, err := anypb.New(user)

	rpcRequest := &pb.RpcRequest{
		ServiceName: "IUserService",
		MethodName:  "GetUser",
		ParamName:   "RpcUser",
		Value:       anyUser,
	}
	i := 0
	fmt.Println(rpcRequest.ParamList[i])

	bytes, _ := proto.Marshal(rpcRequest)

	unmarshal := &pb.RpcRequest{}
	rpcValue := &pb.RpcUser{}
	proto.Unmarshal(bytes, unmarshal)

	fmt.Println("转换的值", unmarshal.Value.UnmarshalTo(rpcValue))
	fmt.Println("转换的value", rpcValue.Name)

	if err != nil {
		return
	}

	//封装一个数据报文
	message := &znet.Message{
		Id:     1,
		Length: uint32(len(bytes)),
		Msg:    bytes,
	}

	packUtils := znet.NewDataPack()
	//封包
	pack, err := packUtils.Pack(message)
	fmt.Println(bytes)
	if err != nil {
		return
	}

	for {
		_, err := conn.Write(pack)
		if err != nil {
			return
		}
		head := make([]byte, 8)
		if _, err := io.ReadFull(conn, head); err != nil {
			return
		}

		//读取完头信息后, 将消息拆包
		message, err := packUtils.Unpack(head)
		if err != nil {
			return
		}

		body := make([]byte, message.GetMsgLen())
		if _, err = io.ReadFull(conn, body); err != nil {
			return
		}
		response := &pb.RpcResponse{}
		proto.Unmarshal(body, response)
		fmt.Println("返回的类型名:", response.TypeName)
		result, err := util.NewTypeRegistry().GetType(response.TypeName)

		response.ResponseValue.UnmarshalTo(result)
		fmt.Println("调用返回的年龄:", result.(*pb.RpcUser).Age)
		message.SetMsg(body)

		//打印出服务器传回来的消息
		fmt.Println("msgId:", message.GetMsgId(), "msgLen:", message.GetMsgLen(), "msgBody", string(message.GetMsg()))
		time.Sleep(2 * time.Second)
	}
}
