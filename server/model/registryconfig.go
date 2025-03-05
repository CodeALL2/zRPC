package model

import "time"

type RegistryConfig struct {
	registry     string //注册中心类别
	registryAddr string //注册中心地址
	userName     string //用户名
	password     string //密码
	timeout      int    //超时时间
}

func NewRegistryConfig(registryName string, registryAddr string) *RegistryConfig {
	return &RegistryConfig{
		registry:     registryName,
		registryAddr: registryAddr,
		timeout:      5,
	}
}

func (s *RegistryConfig) GetRegistryName() string {
	return s.registry
}
func (s *RegistryConfig) SetRegistry(registryName string) {
	s.registry = registryName
}

func (s *RegistryConfig) SetRegistryAddr(registryAddr string) {
	s.registryAddr = registryAddr
}

func (s *RegistryConfig) GetRegistryAddr() string {
	return s.registryAddr
}

func (s *RegistryConfig) GetRegistryTimeout() time.Duration {
	return time.Duration(s.timeout) * time.Second
}
