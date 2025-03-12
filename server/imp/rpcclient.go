package imp

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
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
	serverMapLock  sync.RWMutex
	addr           string
	port           string
	application    *ZRPCApplication
	index          int                         //轮询下标
	serverMap      map[string]net.Conn         //所有的server存储map
	megMap         map[string]*model.MsgResult //所有的消息
	msgHandler     iface.IClientMsgHandle
	serverHeartMap sync.Map //维护心跳的时间
}

func NewRPCClient(application *ZRPCApplication) *RPCClient {
	//需要初始化handler
	c := &RPCClient{
		application: application,
		index:       0,
		serverMap:   make(map[string]net.Conn),
		megMap:      make(map[string]*model.MsgResult),
		msgHandler:  NewClientMsgHandle(),
	}
	c.msgHandler.AddMsgHandler(1, &ClientRPCRouterImp{})
	c.msgHandler.AddMsgHandler(2, &ClientHeartRouterImp{})
	go c.heart() //开启心跳包
	return c
}

func (c *RPCClient) Dial(addr string, port string) (error, net.Conn) {
	con, err := net.Dial("tcp", addr+":"+port)
	if err != nil {
		fmt.Println("连接失败")
		return err, nil
	}
	return nil, con
}

func (c *RPCClient) Invoke(serviceName string, methodName string, paramName string, value interface{}) (*model.MsgResult, error) {

	_, err := util.NewTypeRegistry().GetType(paramName)
	if err != nil {
		fmt.Println("参数类型没有遵循protobuf协议")
		return nil, err
	}
	msgId := uuid.New()
	anyValue, err := anypb.New(value.(proto.Message))
	if err != nil {
		fmt.Println("参数的值没有遵循protobuf协议")
		return nil, err
	}
	rpcRequest := &pb.RpcRequest{
		MsgId:       msgId.String(),
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

	pack, err := c.Pack(bytesData, 1) //rpc默认处理方法
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
	var rpcCon net.Conn
	rpcCon = c.FindCacheServer(metaInfo)
	if rpcCon != nil {
		fmt.Println("服务提供者的缓存连接对象已命中", metaInfo.GetServiceNodeKey())
	} else {
		//与这个服务器进行连接
		err, rpcCon = c.Dial(metaInfo.ServiceHost, metaInfo.ServicePort)
		if err != nil {
			fmt.Println("与服务端服务器连接失败, 清扫本地缓存, 寻找可用连接")
			c.RemoveServerNodeCache(metaInfo.ServiceHost + ":" + metaInfo.ServicePort)
			//找寻下一个可用的服务节点进行兜底
			err, rpcCon = c.FindEtcdClient(registryServer, discovery)
			if err != nil {
				return nil, err
			}
		} else {
			//将连接对象加入到连接池中去
			fmt.Println("服务提供者的缓存连接对象未命中, 已加入连接池", metaInfo.GetServiceNodeKey())
			c.AddServerCacheToMap(metaInfo, rpcCon)
		}
	}
	var result *model.MsgResult

	for {
		result, err = c.sendPackToServer(pack, rpcCon, msgId.String())
		if errors.Is(err, model.ErrClientNotLine) {
			//清扫本地缓存，此服务节点为异常
			c.RemoveServerNodeCache(metaInfo.ServiceHost + ":" + metaInfo.ServicePort)
			// 寻找其他可用节点
			err, rpcCon = c.FindEtcdClient(registryServer, discovery)
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

func (c *RPCClient) Pack(bytesData []byte, msgCode int) ([]byte, error) {
	//封装一个数据报文
	message := &znet.Message{
		Id:     uint32(msgCode),
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

func (c *RPCClient) sendPackToServer(pack []byte, rpcCon net.Conn, msgUUID string) (*model.MsgResult, error) {
	_, err := rpcCon.Write(pack)
	//消息唯一uuid插入
	c.megMap[msgUUID] = &model.MsgResult{
		Result: make(chan interface{}, 1),
	}

	if err != nil {
		fmt.Println("客户端向服务器写入数据失败")
		return nil, model.ErrClientNotLine
	}
	//将这个返回值返回
	return c.megMap[msgUUID], nil
}

func (c *RPCClient) FindEtcdClient(registryServer iface.IRegistryServer, metaInfo []*model.ServiceMetaInfo) (error, net.Conn) {
	cacheLength := len(metaInfo)
	if cacheLength == 0 {
		return errors.New("服务节点列表为空"), nil
	}

	fmt.Printf("开始尝试连接，共有%d个候选节点\n", cacheLength)

	// 可以考虑随机打乱节点顺序，避免总是连接同一个节点
	// rand.Shuffle(len(metaInfo), func(i, j int) { metaInfo[i], metaInfo[j] = metaInfo[j], metaInfo[i] })
	maxLine := 3
	for i, node := range metaInfo {
		nodeAddr := node.ServiceHost + ":" + node.ServicePort
		fmt.Printf("尝试连接节点 %d/%d: %s\n", i+1, cacheLength, nodeAddr)
		//从连接缓存中查询是否有此节点
		cacheClient := c.FindCacheServer(node)
		if cacheClient != nil {
			return nil, cacheClient
		}

		err, rpcCon := c.Dial(node.ServiceHost, node.ServicePort)
		if err == nil {
			fmt.Printf("成功连接到节点: %s 已将其纳入本地节点缓存\n", nodeAddr)
			c.AddServerCacheToMap(node, rpcCon)
			return nil, rpcCon
		}

		fmt.Printf("连接节点失败: %s, 错误: %v\n", nodeAddr, err)

		if i < maxLine && i < cacheLength-1 {
			fmt.Printf("没有正确和服务器连接，正在进行第%d次重试操作寻找可用节点进行连接\n", i+1)
			// 添加短暂延迟，避免频繁重试
			time.Sleep(100 * time.Millisecond)
		} else {
			return errors.New("重试达到最大次数"), nil
		}
		// 清除本地连接缓存
		c.RemoveServerNodeCache(nodeAddr)
	}

	return errors.New("没有发现可用的服务提供者节点"), nil
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
	key := metaInfo.ServiceHost + ":" + metaInfo.ServicePort
	if c.serverMap[key] != nil {
		return
	}

	c.serverMapLock.Lock()
	defer c.serverMapLock.Unlock()
	c.serverMap[key] = conn
	c.serverHeartMap.Store(key, time.Now().Unix()) //将此节点加入心跳维护
	go c.reader(conn, key)                         //在这里开启读协程
}

func (c *RPCClient) RemoveServerNodeCache(serverKey string) {
	c.serverMapLock.Lock()
	defer c.serverMapLock.Unlock()
	delete(c.serverMap, serverKey)
}

func (c *RPCClient) reader(con net.Conn, serverKey string) {
	packUtils := znet.NewDataPack()
	for {
		head := make([]byte, 8)
		if _, err := io.ReadFull(con, head); err != nil {
			fmt.Println("客户端读取头数据失败, 清理本地缓存")
			//清理本地socket缓存 以及心跳组
			c.serverHeartMap.Delete(serverKey)
			c.RemoveServerNodeCache(serverKey)
			return
			//TODO 读取失败应该进行一系列兜底操作
		}

		//读取完头信息后, 将消息拆包
		requestMessage, err := packUtils.Unpack(head)
		if err != nil {
			fmt.Println("客户端解析头数据失败,服务端没有遵循协议要求, 清理本地缓存")
			//TODO 读取失败应该进行一系列兜底操作
			c.serverHeartMap.Delete(serverKey)
			c.RemoveServerNodeCache(serverKey)
			return
		}

		body := make([]byte, requestMessage.GetMsgLen())
		if _, err = io.ReadFull(con, body); err != nil {
			fmt.Println("客户端解析消息体数据失败, 清理本地缓存")
			//TODO 读取失败应该进行一系列兜底操作
			c.serverHeartMap.Delete(serverKey)
			c.RemoveServerNodeCache(serverKey)
			return
		}
		requestMessage.SetMsg(body)

		request := &model.ClientMsgRequest{
			MessageMap:     c.megMap,
			Con:            con,
			Data:           requestMessage,
			HeartServerMap: &c.serverHeartMap,
		}
		//将消息投入handler
		c.msgHandler.SendMessageToClientQueue(request)
	}
}

// 这里可以添加一个回复时间表，如果超过了某一个阈值 就会剔除此节点
func (c *RPCClient) heart() {
	ticker := time.NewTicker(5 * time.Second) // 每 5 秒执行一次
	defer ticker.Stop()

	for range ticker.C { // 每次触发时执行
		for key, value := range c.serverMap {
			timer := false // 时间阈值

			load, ok := c.serverHeartMap.Load(key)
			if !ok {
				continue
			}
			// load 是时间戳
			if timestamp, ok := load.(int64); ok {
				if time.Now().Unix()-timestamp > 15 {
					// 当前时间戳比 load 大 15 秒以上
					fmt.Println("当前节点", key, "触发时间期限阈值")
					timer = true // 将阈值打开
				}
			}

			data := []byte(key)
			message := &znet.Message{
				Id:     2, // 心跳包 id
				Msg:    data,
				Length: uint32(len(data)),
			}
			//// 封包
			pack, err := znet.NewDataPack().Pack(message)
			if err != nil {
				fmt.Println(err)
				return
			}
			_, err = value.Write(pack)
			if err != nil {
				fmt.Println(err)
				// 此节点写入失败 查看时间阈值是否也触发
				if timer {
					// 剔除此节点
					c.serverHeartMap.Delete(key)
					c.RemoveServerNodeCache(key)
					fmt.Println("节点", key, "同时触发时间阈值和次数阈值，剔除此节点")
				}
			}
			fmt.Println("心跳包已发送：", key)
		}
	}
}
