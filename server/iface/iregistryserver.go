package iface

import (
	"zRPC/server/model"
)

type IRegistryServer interface {
	Init(config *model.RegistryConfig) error                              //注册中心的初始化
	Register(info *model.ServiceMetaInfo) error                           //注册服务
	UnRegister(info *model.ServiceMetaInfo) error                         //下架服务
	ServiceDiscovery(serviceKey string) ([]*model.ServiceMetaInfo, error) //返回所有服务
	Destroy() error                                                       //注销注册中心
	HeartBeat(duration int64)                                             //心跳服务
	GetRegistryCache() *model.RegistryCache                               //获取注册中心相关的缓存服务
	WatchKeys()                                                           //监听key
}
