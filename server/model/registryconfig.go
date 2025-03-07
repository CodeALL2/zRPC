package model

type RegistryConfig struct {
	registry     string //注册中心类别
	registryAddr string //注册中心地址
	userName     string //用户名
	password     string //密码
	timeout      int64  //超时时间
}

func NewRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		registry:     "etcd",
		registryAddr: "127.0.0.1:2379",
		timeout:      30,
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

func (s *RegistryConfig) GetRegistryTimeout() int64 {
	return s.timeout
}

func (s *RegistryConfig) SetTimeOut(timeout int64) {
	s.timeout = timeout
}
