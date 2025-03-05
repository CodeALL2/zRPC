package model

/*
*
服务的注册元信息
*/
type ServiceMetaInfo struct {
	ServiceName    string
	ServiceVersion string
	ServiceHost    string
	ServicePort    string
}

/*
*
获取服务的前缀名
*/
func (s *ServiceMetaInfo) GetServiceKey() string {
	return s.ServiceName + ":" + s.ServiceVersion
}

/*
*
获取服务的整体名称
*/
func (s *ServiceMetaInfo) GetServiceNodeKey() string {
	return s.GetServiceKey() + "/" + s.ServiceHost + ":" + s.ServicePort
}
