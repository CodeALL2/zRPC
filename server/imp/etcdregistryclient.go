package imp

import (
	"context"
	"encoding/json"
	"fmt"
	clientV3 "go.etcd.io/etcd/client/v3"
	"log"
	"regexp"
	"time"
	"zRPC/server/model"
)

type EtcdRegistryClient struct {
	BaseRegistryServer
	client        *clientV3.Client
	registryCache *model.RegistryCache
}

func NewEtcdRegistryClient() *EtcdRegistryClient {
	return &EtcdRegistryClient{
		client:        nil,
		registryCache: model.NewRegistryCache(),
	}
}

func (s *EtcdRegistryClient) Init(config *model.RegistryConfig) error { // 注册中心的初始化
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
	//客户端不需要心跳检测
	//timer := time.NewTicker(time.Duration(config.GetRegistryTimeout()/2) * time.Second)
	//go func() {
	//	fmt.Println("心跳检测已启动")
	//	heartHz := config.GetRegistryTimeout()
	//	for {
	//		select {
	//		case <-timer.C:
	//			s.HeartBeat(int64(heartHz))
	//		}
	//	}
	//}()

	//监听器
	fmt.Println("开启一个监听器")
	s.WatchKeys()
	return nil
}

func (s *EtcdRegistryClient) ServiceDiscovery(serviceKey string) ([]*model.ServiceMetaInfo, error) { //获取所有注册服务
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

func (s *EtcdRegistryClient) GetRegistryCache() *model.RegistryCache {
	return s.registryCache
}

// 客户端专属
func (s *EtcdRegistryClient) WatchKeys() { //监听键值
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
					// 清空本地缓存
					re := regexp.MustCompile(`(/rpc/[^/]+)`)
					matches := re.FindStringSubmatch(key)
					if len(matches) > 0 {
						prefix := matches[1]
						fmt.Println(prefix) // 输出: /rpc/IUserService:v1.0
						s.flushCache(prefix)
					}
				}
			}
		}
	}()
}

// 客户端专属
func (s *EtcdRegistryClient) flushCache(key string) {
	s.registryCache.FlushMateInfoCache(key) //将本地的key清空
}
