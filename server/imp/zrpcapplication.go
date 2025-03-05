package imp

import (
	"zRPC/server/iface"
	"zRPC/server/model"
)

type ZRPCApplication struct {
	registryConfig *model.RegistryConfig
	regisFactory   *RegistryFactory
}

func NewZRPCApplication() *ZRPCApplication {
	registryFactory := NewRegistryFactory()
	return &ZRPCApplication{
		registryConfig: &model.RegistryConfig{},
		regisFactory:   registryFactory,
	}
}

func (z *ZRPCApplication) GetRegistryConfig() *model.RegistryConfig {
	return z.registryConfig
}

func (z *ZRPCApplication) GetRegistryFactory() iface.IRegisFactory {
	return z.regisFactory
}
