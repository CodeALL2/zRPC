package ziface

type IConnManager interface {
	AddConnection(conn IConnection)          //添加链接
	RemoveConnection(conn IConnection)       //删除链接
	GetConnection(connID uint32) IConnection //获取连接
	ConnectionLen() uint32                   //当前连接的总数
	ClearConnection()                        //关闭并且清除所有的链接
}
