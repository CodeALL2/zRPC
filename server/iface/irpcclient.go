package iface

type IRPCClient interface {
	Dial() error
	Invoke(serviceName string, methodName string, paramName string, value interface{}) error
}
