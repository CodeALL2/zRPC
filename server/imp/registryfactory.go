package imp

import (
	"zRPC/server/iface"
)

type RegistryFactory struct {
	registryMap map[string]iface.IRegistryServer
}

func NewRegistryFactory() *RegistryFactory {
	registryFactory := &RegistryFactory{
		registryMap: make(map[string]iface.IRegistryServer),
	}
	registryFactory.InitRegistryServer()
	return registryFactory
}

func (r *RegistryFactory) GetRegistryServer(name string) iface.IRegistryServer {
	return r.registryMap[name]
}

func (r *RegistryFactory) InitRegistryServer() {
	r.registryMap["etcd-server"] = NewEtcdRegistryServer()
	r.registryMap["etcd-client"] = NewEtcdRegistryClient()
}
