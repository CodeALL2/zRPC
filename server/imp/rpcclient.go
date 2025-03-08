package imp

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net"
	"sync"
	"time"
	"zRPC/gzinx/znet"
	"zRPC/protobufdemo/pp/pb"
	"zRPC/server/iface"
	"zRPC/server/model"
	"zRPC/server/util"
)

type RPCClient struct {
	serverMapLock sync.RWMutex
	con           net.Conn
	addr          string
	port          string
	application   *ZRPCApplication
	index         int                 //轮询下标
	serverMap     map[string]net.Conn //所有的server存储map
}

func NewRPCClient(application *ZRPCApplication) *RPCClient {
	return &RPCClient{
		application: application,
		index:       0,
		serverMap:   make(map[string]net.Conn),
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
	clientCache := c.FindCacheServer(metaInfo)

	if clientCache != nil {
		c.con = clientCache
		fmt.Println("服务提供者的缓存连接对象已命中", metaInfo.GetServiceNodeKey())
	} else {
		//与这个服务器进行连接
		err = c.Dial(metaInfo.ServiceHost, metaInfo.ServicePort)
		if err != nil {
			fmt.Println("与服务端服务器连接失败, 清扫本地缓存, 寻找可用连接")
			c.RemoveServerNodeCache(metaInfo.ServiceHost + ":" + metaInfo.ServicePort)
			//找寻下一个可用的服务节点进行兜底
			err := c.FindEtcdClient(registryServer, discovery)
			if err != nil {
				return nil, err
			}
		} else {
			//将连接对象加入到连接池中去
			fmt.Println("服务提供者的缓存连接对象未命中, 已加入连接池", metaInfo.GetServiceNodeKey())
			c.AddServerCacheToMap(metaInfo, c.con)
		}
	}
	var result interface{}

	for {
		result, err = c.sendPackToServer(pack)
		if errors.Is(err, model.ErrClientNotLine) {
			// 寻找其他可用节点
			err := c.FindEtcdClient(registryServer, discovery)
			if err != nil {
				fmt.Println("没有搜寻到可用节点")
				// 继续循环尝试，直到达到最大重试次数
				return nil, err
			}
		} else { // 如果连接返回没有错误 或 错误不是连接失败导致的就直接跳出
			break
		}
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
		return nil, model.ErrClientNotLine
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
	cacheLength := len(metaInfo)
	if cacheLength == 0 {
		return errors.New("服务节点列表为空")
	}

	fmt.Printf("开始尝试连接，共有%d个候选节点\n", cacheLength)

	// 可以考虑随机打乱节点顺序，避免总是连接同一个节点
	// rand.Shuffle(len(metaInfo), func(i, j int) { metaInfo[i], metaInfo[j] = metaInfo[j], metaInfo[i] })
	maxLine := 3
	for i, node := range metaInfo {
		nodeAddr := node.ServiceHost + ":" + node.ServicePort
		fmt.Printf("尝试连接节点 %d/%d: %s\n", i+1, cacheLength, nodeAddr)

		err := c.Dial(node.ServiceHost, node.ServicePort)
		if err == nil {
			fmt.Printf("成功连接到节点: %s 已将其纳入本地节点缓存\n", nodeAddr)
			c.AddServerCacheToMap(node, c.con)
			return nil
		}

		fmt.Printf("连接节点失败: %s, 错误: %v\n", nodeAddr, err)

		if i < maxLine && i < cacheLength-1 {
			fmt.Printf("没有正确和服务器连接，正在进行第%d次重试操作寻找可用节点进行连接\n", i+1)
			// 添加短暂延迟，避免频繁重试
			time.Sleep(100 * time.Millisecond)
		} else {
			return errors.New("重试达到最大次数")
		}
		// 清除本地连接缓存
		c.RemoveServerNodeCache(nodeAddr)
	}

	return errors.New("没有发现可用的服务提供者节点")
}

func (c *RPCClient) FindCacheServer(metaInfo *model.ServiceMetaInfo) net.Conn {
	c.serverMapLock.RLock()
	defer c.serverMapLock.RUnlock()
	serverCacheKey := metaInfo.ServiceHost + ":" + metaInfo.ServicePort
	clientCache := c.serverMap[serverCacheKey]
	if clientCache != nil {
		return clientCache
	}
	return nil
}

func (c *RPCClient) AddServerCacheToMap(metaInfo *model.ServiceMetaInfo, conn net.Conn) {
	c.serverMapLock.Lock()
	defer c.serverMapLock.Unlock()
	key := metaInfo.ServiceHost + ":" + metaInfo.ServicePort
	c.serverMap[key] = conn
}

func (c *RPCClient) RemoveServerNodeCache(serverKey string) {
	c.serverMapLock.Lock()
	defer c.serverMapLock.Unlock()
	delete(c.serverMap, serverKey)
}
