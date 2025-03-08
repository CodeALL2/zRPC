package znet

import (
	"fmt"
	"net"
	"zRPC/gzinx/utils"
	"zRPC/gzinx/ziface"
)

type Server struct {
	//ip
	IP string
	//端口
	Port int32
	//服务器的名称
	Name string
	//服务器的ip版本
	IPVersion string
	//用户定义的路由
	msgHandler ziface.IMsgHandle
	//连接管理
	connMgr ziface.IConnManager
	//连接后进行的方法调用
	onConnStartFunc func(conn ziface.IConnection)
	//连接关闭之前进行的调用
	onConnStopFunc func(conn ziface.IConnection)
}

func NewServer(name string) ziface.IServer {
	s := &Server{
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.Port,
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp",
		msgHandler: nil,
		connMgr:    NewConnManager(),
	}
	return s
}

func (s *Server) AddMsgHandler(msgHandler ziface.IMsgHandle) {
	s.msgHandler = msgHandler
}

// 对接口的实现
func (s *Server) Start() {
	fmt.Println("[gzinx server] ", "name:", s.Name, " ip", s.IP, " port:", s.Port)
	fmt.Println("[gzinx] ", "version:", utils.GlobalObject.Version, " maxconn", utils.GlobalObject.MaxConn)
	go func() {
		listen, err := net.Listen(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		defer listen.Close()

		if err != nil {
			fmt.Println("tcp and port error:", err)
			return
		}

		s.msgHandler.StartWorkerPoll()

		fmt.Println("gzinx start success ip:", s.IP, "port:", s.Port)
		var connId uint32 = 0
		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Println("Accept error:", err)
				continue
			}

			//判断是否已经达到连接的最大总数
			if s.connMgr.ConnectionLen() >= uint32(utils.GlobalObject.MaxConn) {
				//在这里要给客户端发一个链接已经溢出的包
				conn.Close()
				fmt.Println("[gzinx] tcp server over max conn")
				continue
			}

			connId++
			connection := NewConnection(s, conn, connId, s.msgHandler)
			go connection.Star()
		}
	}()
}

func (s *Server) Stop() {
	//终止所有的链接
	fmt.Println("服务器正在关闭中...")
	s.connMgr.ClearConnection()
	fmt.Println("服务器已关闭")
}

func (s *Server) Serve() {
	s.Start()
	//启动服务器后做一些其他的业务操作
	select {}
}

func (s *Server) GetConnManager() ziface.IConnManager {
	return s.connMgr
}

// 添加hook接口
func (s *Server) AddOnConnStartHook(startHook func(connection ziface.IConnection)) {
	s.onConnStartFunc = startHook
}
func (s *Server) AddOnConnStopHook(stopHook func(connection ziface.IConnection)) {
	s.onConnStopFunc = stopHook
}

// 调用接口
func (s *Server) CallConnStartHook(connection ziface.IConnection) {
	if s.onConnStartFunc != nil {
		fmt.Println("[gzinx] call conn start hook ok")
		s.onConnStartFunc(connection)
	}
}

func (s *Server) CallConnStopHook(connection ziface.IConnection) {
	if s.onConnStopFunc != nil {
		fmt.Println("[gzinx] call conn stop hook ok")
		s.onConnStopFunc(connection)
	}
}
