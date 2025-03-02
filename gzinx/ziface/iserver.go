package ziface

type IServer interface {
	//服务器的开启
	Start()
	//服务器的关闭
	Stop()
	//服务器的运行
	Serve()
	//添加用户自定义接口
	AddMsgHandler(msgHandler IMsgHandle)
	//获取连接管理的接口
	GetConnManager() IConnManager
	//添加hook接口
	AddOnConnStartHook(func(connection IConnection))
	AddOnConnStopHook(func(connection IConnection))
	//调用接口
	CallConnStartHook(connection IConnection)
	CallConnStopHook(connection IConnection)
}
