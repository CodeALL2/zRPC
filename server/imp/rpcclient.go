package imp

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/iface"
	"zRPC/server/model"
	"zRPC/server/util"
)

type RPCClient struct {
	con         net.Conn
	addr        string
	port        string
	application *ZRPCApplication
	index       int //轮询下标

}

func NewRPCClient(application *ZRPCApplication) *RPCClient {
	return &RPCClient{
		application: application,
		index:       0,
	}
}

func (c *RPCClient) Dial(addr string, port string) error {
	con, err := net.Dial("tcp", addr+":"+port)
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
		fmt.Println("参数的值没有遵循protobuf协议")
		return nil, err
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
	//在这里选择一个服务端调用
	registryConfig := c.application.GetRegistryConfig()
	registryServer := c.application.GetRegistryFactory().GetRegistryServer(registryConfig.GetRegistryName())
	err = registryServer.Init(registryConfig)
	if err != nil {
		return nil, err
	}

	serviceMetaInfo := &model.ServiceMetaInfo{
		ServiceName:    serviceName,
		ServiceVersion: "v1.0",
	}
	discovery, err := registryServer.ServiceDiscovery(serviceMetaInfo.GetServiceKey())
	if err != nil {
		fmt.Println("获取服务列表失败")
		return nil, err
	}
	length := len(discovery)
	if length == 0 {
		fmt.Println("注册中心没有发现此方法")
		return nil, err
	}

	metaInfo := discovery[c.index%length] //采用轮询算法
	c.index++
	//获取缓存的连接池对象
	clientCache := registryServer.GetRegistryCache().ReadCacheFromServerClientCache(metaInfo.ServiceHost + metaInfo.ServicePort)
	if clientCache != nil {
		c.con = clientCache
		fmt.Println("服务提供者的缓存连接对象已命中", metaInfo.GetServiceNodeKey())

	} else {
		//与这个服务器进行连接
		err = c.Dial(metaInfo.ServiceHost, metaInfo.ServicePort)
		if err != nil {
			fmt.Println("与服务端服务器连接失败, 清扫本地缓存, 寻找可用连接")
			//找寻下一个可用的进行兜底
			err := c.FindEtcdClient(registryServer, discovery)
			if err != nil {
				return nil, err
			}
		} else {
			//将连接对象加入到连接池中去
			fmt.Println("服务提供者的缓存连接对象未命中, 已加入连接池", metaInfo.GetServiceNodeKey())
			registryServer.GetRegistryCache().WriteCacheToServerClientCache(metaInfo.ServiceHost+metaInfo.ServicePort, c.con)
		}
	}

	result, err := c.sendPackToServer(pack)
	if err != nil {
		//TODO  这里应该加入兜底操作，立马切换剩下的可用节点再次尝试
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

func (c *RPCClient) FindEtcdClient(registryServer iface.IRegistryServer, metaInfo []*model.ServiceMetaInfo) error {

	for i := 0; i < len(metaInfo); i++ {

		err := c.Dial(metaInfo[i].ServiceHost, metaInfo[i].ServicePort)
		if err == nil {
			return nil
		}
		//清扫本地连接缓存
		registryServer.GetRegistryCache().FlushServerClientCache(metaInfo[i].ServiceHost + metaInfo[i].ServicePort)
	}
	return errors.New("没有发现可用的服务提供者节点")
}
