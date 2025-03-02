package iface

import (
	"zRPC/server/model"
)

type IRegistry interface {
	LocalRegistry(serviceName string, service interface{}) error //注册器
	RemoteRegistry(serviceName string) error                     //删除注册器
	GetService(serviceName string) (*model.ServiceInfoo, error)  //获取某一个service
}
