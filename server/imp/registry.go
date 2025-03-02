package imp

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"zRPC/server/model"
)

type Registry struct {
	//map
	registryMap sync.Map //存储注册方法的map
}

func NewRegistry() *Registry {
	return &Registry{
		registryMap: sync.Map{},
	}
}
func (rg *Registry) LocalRegistry(serviceName string, service interface{}) error { //注册器
	if service == nil {
		log.Println("注册失败，注册实体为空")
		return errors.New("service is nil")
	}
	//通过反射获取值
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service)

	if serviceType.Kind() != reflect.Ptr {
		log.Println("注册的service必须要求是指针")
		return errors.New("service not ptr")
	}

	serviceInfo := &model.ServiceInfoo{
		ServiceValue: serviceValue,
		ServiceType:  reflect.TypeOf(serviceValue),
	}
	rg.registryMap.Store(serviceName, serviceInfo)
	return nil
}
func (rg *Registry) RemoteRegistry(serviceName string) error { //删除注册器
	rg.registryMap.Delete(serviceName)
	return nil
}
func (rg *Registry) GetService(serviceName string) (*model.ServiceInfoo, error) { //获取某一个service
	if value, ok := rg.registryMap.Load(serviceName); ok {
		serviceInfo := value.(*model.ServiceInfoo)
		return serviceInfo, nil
	}
	return nil, errors.New("没有找到此service")
}
