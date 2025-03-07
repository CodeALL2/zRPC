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
	client               *clientV3.Client
	localRegistryNodeKey map[string]string
	registryCache        *model.RegistryCache
}

const ETCD_ROOT_PAHT = "/rpc/"

func NewEtcdRegistryServer() *EtcdRegistryServer {
	return &EtcdRegistryServer{
		client:               nil,
		localRegistryNodeKey: make(map[string]string),
		registryCache:        model.NewRegistryCache(),
	}
}

func (s *EtcdRegistryServer) Init(config *model.RegistryConfig) error { // 注册中心的初始化
	if s.client != nil {
		return nil
	}

	client, err := clientV3.New(clientV3.Config{
		Endpoints:   []string{config.GetRegistryAddr()},
		DialTimeout: time.Duration(config.GetRegistryTimeout()/2) * time.Second,
	})
	if err != nil {
		fmt.Println("连接不到etcd:", config.GetRegistryAddr())
		return err
	}
	s.client = client
	timer := time.NewTicker(time.Duration(config.GetRegistryTimeout()/2) * time.Second)
	go func() {
		fmt.Println("心跳检测已启动")
		heartHz := config.GetRegistryTimeout()
		for {
			select {
			case <-timer.C:
				s.HeartBeat(int64(heartHz))
			}
		}
	}()
	//监听器
	fmt.Println("开启一个监听器")
	s.WatchKeys()
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
	s.localRegistryNodeKey[registryKey] = string(registryValue) //绑定到本地

	if err != nil {
		fmt.Println("put key to etcd error")
		return err
	}

	return nil
}

func (s *EtcdRegistryServer) UnRegister(info *model.ServiceMetaInfo) error { // 下架服务
	registryKey := ETCD_ROOT_PAHT + info.GetServiceNodeKey()
	_, err := s.client.Delete(context.Background(), ETCD_ROOT_PAHT+info.GetServiceNodeKey())
	if err != nil {
		fmt.Println("删除注册key失败")
		return err
	}
	//从本地删除
	delete(s.localRegistryNodeKey, registryKey)
	return nil
}

func (s *EtcdRegistryServer) ServiceDiscovery(serviceKey string) ([]*model.ServiceMetaInfo, error) { //获取所有注册服务
	//创建前缀key
	prefixKey := ETCD_ROOT_PAHT + serviceKey
	//优先查询本地缓存
	mateInfoCache := s.registryCache.ReadCacheFromMateInfoCache(prefixKey)
	if mateInfoCache != nil {
		fmt.Println("获取注册中心的信息 缓存命中 key:", prefixKey)
		return mateInfoCache, nil
	}

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
	//写入本地缓存
	fmt.Println("将注册中心的信息缓存到本地 key:", prefixKey)
	s.registryCache.WriteCacheToMateInfoCache(prefixKey, services)
	return services, nil
}

func (s *EtcdRegistryServer) Destroy() error { //注销注册中心

	for key, _ := range s.localRegistryNodeKey {
		_, err := s.client.Delete(context.Background(), key, clientV3.WithPrefix())
		if err != nil {
			fmt.Println("删除etcd", key, "失败")
			continue
		}
	}

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

func (s *EtcdRegistryServer) HeartBeat(heartHz int64) { //心跳检测
	fmt.Println("心跳检测中")
	for registryKey, _ := range s.localRegistryNodeKey {
		resp, err := s.client.Get(context.Background(), registryKey)

		if err != nil {
			fmt.Println("续期心跳连接服务器失败", err)
			continue
		}

		if len(resp.Kvs) == 0 {
			fmt.Println("服务", registryKey, "已经过期, 正在续期")
			lease, err := s.client.Grant(context.Background(), int64(heartHz)) //创建一个续约
			if err != nil {
				fmt.Println("创建续期失败", registryKey)
				continue
			}
			_, err = s.client.Put(context.Background(), registryKey, string(resp.Kvs[0].Value), clientV3.WithLease(lease.ID))
			if err != nil {
				fmt.Println("重新注册续期失败", registryKey)
				continue
			}
			fmt.Println("服务器信息重新注册成功", registryKey)
		} else { //正常续期
			fmt.Println("服务正在续期中", registryKey)
			lease, err := s.client.Grant(context.Background(), int64(heartHz))
			if err != nil {
				fmt.Println("服务创建续期失败", registryKey)
				continue
			}
			_, err = s.client.Put(context.Background(), registryKey, string(resp.Kvs[0].Value), clientV3.WithLease(lease.ID))
			if err != nil {
				fmt.Println("服务续期失败", registryKey)
				continue
			}
			fmt.Println("服务器信息重新注册成功", registryKey)
		}
	}
}

func (s *EtcdRegistryServer) GetRegistryCache() *model.RegistryCache {
	return s.registryCache
}

func (s *EtcdRegistryServer) WatchKeys() { //监听键值
	watchChan := s.client.Watch(context.Background(), ETCD_ROOT_PAHT, clientV3.WithPrefix())

	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				key := string(event.Kv.Key)
				value := string(event.Kv.Value)

				switch event.Type {
				case clientV3.EventTypePut: // 新增或更新
					fmt.Printf("键 %s 已更新，新值: %s\n", key, value)
				case clientV3.EventTypeDelete: // 删除
					fmt.Printf("键 %s 已删除\n", key)
					var metaInfo = &model.ServiceMetaInfo{}
					fmt.Println("值：", value)
					json.Unmarshal([]byte(value), metaInfo)
					// 清空本地缓存
					s.flushCache(metaInfo)
				}
			}
		}
	}()
}

func (s *EtcdRegistryServer) flushCache(info *model.ServiceMetaInfo) {
	prefixKey := ETCD_ROOT_PAHT + info.GetServiceKey()
	s.registryCache.FlushMateInfoCache(prefixKey, info) //将本地的key清空
}
