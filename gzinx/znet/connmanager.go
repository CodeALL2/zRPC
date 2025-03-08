package znet

import (
	"fmt"
	"sync"
	"zRPC/gzinx/ziface"
)

type ConnManager struct {
	Connections map[uint32]ziface.IConnection //存放所有的链接数
	connLock    sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		Connections: make(map[uint32]ziface.IConnection),
	}
}

func (manager *ConnManager) AddConnection(conn ziface.IConnection) { //添加链接
	manager.connLock.Lock()
	defer manager.connLock.Unlock()
	id := conn.GetConnID()
	if manager.Connections[id] != nil {
		fmt.Println("链接添加失败，已存在", id)
		return
	}
	manager.Connections[id] = conn
	fmt.Println("链接添加成功", id)
}
func (manager *ConnManager) RemoveConnection(conn ziface.IConnection) { //删除链接
	manager.connLock.Lock()
	defer manager.connLock.Unlock()
	id := conn.GetConnID()
	delete(manager.Connections, id)
	fmt.Println("链接删除成功", id)
}

func (manager *ConnManager) GetConnection(connID uint32) ziface.IConnection { //获取连接
	manager.connLock.RLock()
	defer manager.connLock.RUnlock()
	if conn, ok := manager.Connections[connID]; ok {
		return conn
	}
	return nil
}
func (manager *ConnManager) ConnectionLen() uint32 { //当前连接的总数
	manager.connLock.RLock()
	defer manager.connLock.RUnlock()
	return uint32(len(manager.Connections))
}
func (manager *ConnManager) ClearConnection() { //关闭并且清除所有的链接
	manager.connLock.Lock()
	defer manager.connLock.Unlock()
	for id, conn := range manager.Connections {
		conn.Stop()
		delete(manager.Connections, id)
	}
	fmt.Println("所有的链接已经删除完毕")
}
