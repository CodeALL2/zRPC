package imp

import (
	"fmt"
	"zRPC/server/model"
)

type BaseRegistryServer struct {
}

func (b *BaseRegistryServer) Init(config *model.RegistryConfig) error { //注册中心的初始化
	fmt.Println("请实现各自的init")
	return nil
}
func (b *BaseRegistryServer) Register(info *model.ServiceMetaInfo) error { //注册服务
	fmt.Println("请实现各自的register")
	return nil
}
func (b *BaseRegistryServer) UnRegister(info *model.ServiceMetaInfo) error { //下架服务
	fmt.Println("请实现各自的unregister")
	return nil
}
func (b *BaseRegistryServer) ServiceDiscovery(serviceKey string) ([]*model.ServiceMetaInfo, error) { //返回所有服务
	fmt.Println("请实现各自的serviceDiscovery")
	return nil, nil
}

func (b *BaseRegistryServer) Destroy() error { //注销注册中心
	fmt.Println("请实现各自的Discovery")
	return nil
}
func (b *BaseRegistryServer) HeartBeat(duration int64) { //心跳服务
	fmt.Println("请实现各自的HeartBeat")
}
func (b *BaseRegistryServer) GetRegistryCache() *model.RegistryCache { //获取注册中心相关的缓存服务
	fmt.Println("请实现各自的GetRegistryCache")
	return nil
}

func (b *BaseRegistryServer) WatchKeys() {
	fmt.Println("请实现各自的WatchKeys")
}
