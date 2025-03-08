package iface

import (
	"zRPC/server/model"
)

type IIZRPCApplication interface {
	GetRegistryConfig() *model.RegistryConfig
	GetRegistryFactory() IRegisFactory
}
