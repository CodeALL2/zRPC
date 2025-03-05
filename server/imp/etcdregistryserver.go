package imp

import (
	"context"
	"encoding/json"
	"fmt"
	clientV3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
	"zRPC/server/model"
)

type EtcdRegistryServer struct {
	client *clientV3.Client
}

const ETCD_ROOT_PAHT = "/rpc/"

func NewEtcdRegistryServer() *EtcdRegistryServer {
	return &EtcdRegistryServer{
		client: nil,
	}
}

func (s *EtcdRegistryServer) Init(config *model.RegistryConfig) error { // 注册中心的初始化

	client, err := clientV3.New(clientV3.Config{
		Endpoints:   []string{config.GetRegistryAddr()},
		DialTimeout: config.GetRegistryTimeout() * time.Second,
	})
	if err != nil {
		fmt.Println("连接不到etcd:", config.GetRegistryAddr())
		return err
	}
	s.client = client
	return nil
}

func (s *EtcdRegistryServer) Register(info *model.ServiceMetaInfo) error { // 注册服务
	lease, err := s.client.Grant(context.Background(), 30) //创建一个续约
	if err != nil {
		fmt.Println("创建租约失败")
		return err
	}

	//创建一个etcd key
	registryKey := ETCD_ROOT_PAHT + info.GetServiceNodeKey()
	//创建一个etcd value
	registryValue, err := json.Marshal(info)
	if err != nil {
		fmt.Println("注册中心元数据转json失败")
		return err
	}

	//将key value塞入 并绑定到创建的续约上
	_, err = s.client.Put(context.Background(), registryKey, string(registryValue), clientV3.WithLease(lease.ID))

	if err != nil {
		fmt.Println("put key to etcd error")
		return err
	}

	return nil
}

func (s *EtcdRegistryServer) UnRegister(info *model.ServiceMetaInfo) error { // 下架服务
	_, err := s.client.Delete(context.Background(), ETCD_ROOT_PAHT+info.GetServiceNodeKey())
	if err != nil {
		fmt.Println("删除注册key失败")
		return err
	}
	return nil
}

func (s *EtcdRegistryServer) ServiceDiscovery(serviceKey string) ([]*model.ServiceMetaInfo, error) { //获取所有注册服务
	//创建前缀key
	prefixKey := ETCD_ROOT_PAHT + serviceKey
	//创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//使用前缀查询
	resp, err := s.client.Get(ctx, prefixKey, clientV3.WithPrefix())
	if err != nil {
		log.Printf("查询服务失败: %v", err)
		return nil, err
	}
	// 解析结果
	services := make([]*model.ServiceMetaInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		service := &model.ServiceMetaInfo{}
		err := json.Unmarshal(kv.Value, service)
		if err != nil {
			log.Printf("解析服务信息失败: %v", err)
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *EtcdRegistryServer) Destroy() error { //注销注册中心
	if s.client == nil {
		return nil
	}
	err := s.client.Close()
	if err != nil {
		fmt.Println("注销注册中心失败")
		return err
	}
	return nil
}
