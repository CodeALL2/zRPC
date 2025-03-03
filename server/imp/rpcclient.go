package imp

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/util"
)

type RPCClient struct {
	con  net.Conn
	addr string
	port string
}

func NewRPCClient(addr string, port string) *RPCClient {
	return &RPCClient{
		con:  nil,
		addr: addr,
		port: port,
	}
}

func (c *RPCClient) Dial() error {
	con, err := net.Dial("tcp", c.addr+":"+c.port)
	if err != nil {
		fmt.Println("连接失败")
		return err
	}
	c.con = con
	return nil
}

func (c *RPCClient) Invoke(serviceName string, methodName string, paramName string, value interface{}) (interface{}, error) {
	_, err := util.NewTypeRegistry().GetType(paramName)
	if err != nil {
		fmt.Println("参数类型没有遵循protobuf协议")
		return nil, err
	}

	anyValue, err := anypb.New(value.(proto.Message))
	if err != nil {
		return nil, err
		fmt.Println("参数的值没有遵循protobuf协议")
	}
	rpcRequest := &pb.RpcRequest{
		ServiceName: serviceName,
		MethodName:  methodName,
		ParamName:   paramName,
		Value:       anyValue,
	}

	bytesData, err := proto.Marshal(rpcRequest)

	if err != nil {
		fmt.Println("protoBuff协议转换失败", err)
		return nil, err
	}

	pack, err := c.Pack(bytesData)
	if err != nil {
		fmt.Println("封包失败")
		return nil, err
	}
	result, err := c.sendPackToServer(pack)
	if err != nil {
		fmt.Println("没有正确获取到服务器返回的protobuf数据")
		return nil, err
	}

	return result, err
}

func (c *RPCClient) Pack(bytesData []byte) ([]byte, error) {
	//封装一个数据报文
	message := &znet.Message{
		Id:     1,
		Length: uint32(len(bytesData)),
		Msg:    bytesData,
	}
	packUtils := znet.NewDataPack()
	//封包
	bytePack, err := packUtils.Pack(message)
	if err != nil {
		return nil, err
	}
	return bytePack, nil
}

func (c *RPCClient) sendPackToServer(pack []byte) (interface{}, error) {
	packUtils := znet.NewDataPack()
	_, err := c.con.Write(pack)
	if err != nil {
		fmt.Println("客户端向服务器写入数据失败")
		return nil, err
	}

	head := make([]byte, 8)
	if _, err := io.ReadFull(c.con, head); err != nil {
		fmt.Println("客户端读取头数据失败")
		return nil, err
	}

	//读取完头信息后, 将消息拆包
	message, err := packUtils.Unpack(head)
	if err != nil {
		fmt.Println("客户端解析头数据失败,服务端没有遵循协议要求")
		return nil, err
	}

	body := make([]byte, message.GetMsgLen())
	if _, err = io.ReadFull(c.con, body); err != nil {
		fmt.Println("客户端解析消息体数据失败")
		return nil, err
	}

	//body就是服务端返回的信息
	response := &pb.RpcResponse{}
	if err = proto.Unmarshal(body, response); err != nil {
		fmt.Println("protobuf解析服务器返回的消息失败")
		return nil, err
	}

	fmt.Println("返回的类型名:", response.TypeName)

	result, err := util.NewTypeRegistry().GetType(response.TypeName)
	if err != nil {
		fmt.Println("传回来的协议没有遵循protobuf")
		return nil, err
	}

	if err = response.ResponseValue.UnmarshalTo(result); err != nil {
		fmt.Println("客户端解析消息protobuf返回结果数据失败")
		return nil, err
	}

	return result, nil
}
