package imp

import (
	"fmt"
	"zRPC/gzinx/utils"
	"zRPC/gzinx/znet"
	"zRPC/server/iface"
	"zRPC/server/util"
)

type Server struct {
	Registry       iface.IRegistry       //registry注册器
	RegistryServer iface.IRegistryServer //注册中心的连接器
}

func NewServer() iface.IServer {
	return &Server{
		Registry: nil,
	}
}

func (s *Server) Start() {
	if util.ConfigPath == "" {
		fmt.Println("使用默认配置")
	} else {
		utils.Reload(util.ConfigPath) //路径配置好
	}

	zRPCServer := znet.NewServer("zRPC")
	//需要初始话handler
	handle := znet.NewMsgHandle()
	handle.AddMsgHandler(1, &TwoHandler{Server: s})   //处理rpc请求的handler
	handle.AddMsgHandler(2, &HeartHandler{Server: s}) //处理心跳的handler
	zRPCServer.AddMsgHandler(handle)
	zRPCServer.Serve()
	//整个函数结束了需要关闭相应资源
	s.CloseRegistryServer()
}

func (s *Server) SetRegistry(registry iface.IRegistry) {
	s.Registry = registry
}
func (s *Server) SetRegistryServer(registry iface.IRegistryServer) {
	s.RegistryServer = registry
}

func (s *Server) CloseRegistryServer() {
	fmt.Println("销毁资源中")
	s.RegistryServer.Destroy()
}

func (s *Server) GetRegistry() iface.IRegistry {
	return s.Registry
}
