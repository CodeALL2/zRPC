package iface

type IRegisFactory interface {
	GetRegistryServer(name string) IRegistryServer
	InitRegistryServer()
}
