package iface

type IServer interface {
	Start()
	SetRegistry(registry IRegistry)
	SetRegistryServer(registry IRegistryServer)
	GetRegistry() IRegistry
}
