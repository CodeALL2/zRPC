package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zRPC/gzinx/ziface"
)

type Connection struct {
	Conn         net.Conn               //当前连接的套接字
	ConnID       uint32                 //当前的链接id
	isClosed     bool                   //当前连接是否关闭
	ExitChan     chan bool              //等待连接关闭的管道
	MsgHandler   ziface.IMsgHandle      //用户自定义的钩子函数
	Pack         ziface.IDataPack       //封包拆包工具
	msgChan      chan []byte            //读写分离的管道
	server       ziface.IServer         //所属的server
	property     map[string]interface{} //属性池
	propertyLock sync.RWMutex           //属性池的读写锁
}

func NewConnection(server ziface.IServer, conn net.Conn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	con := &Connection{
		Conn:       conn,
		ConnID:     connID,
		ExitChan:   make(chan bool, 1),
		isClosed:   false,
		MsgHandler: msgHandler,
		Pack:       NewDataPack(),
		msgChan:    make(chan []byte),
		server:     server,
	}
	//将当前的链接加入到server当中去
	server.GetConnManager().AddConnection(con)
	return con
}

func (c *Connection) StarReader() { //专门负责读
	fmt.Println("Reader Goroutine is running", c.ConnID)
	defer fmt.Println("Reader goroutine is stopped", c.ConnID)
	defer c.Stop()

	for {
		buf := make([]byte, c.Pack.GetPackHeadLen())
		//读取包头
		if _, err := io.ReadFull(c.Conn, buf); err != nil {
			fmt.Println("read msg head error:", err)
			break
		}

		//拆包
		msgHead, err := c.Pack.Unpack(buf)
		if err != nil {
			fmt.Println("unpack error:", err)
			break
		}

		//读取包体
		msgBody := make([]byte, msgHead.GetMsgLen())
		if msgHead.GetMsgLen() > 0 {
			if _, err := io.ReadFull(c.Conn, msgBody); err != nil {
				fmt.Println("read msg body error:", err)
				break
			}
		}

		msgHead.SetMsg(msgBody)

		//封装一个request
		request := &Request{
			conn: c,
			data: msgHead,
		}

		////开始启动调用的业务
		//丢给工作池
		if c.MsgHandler.GetWorkPoolSize() > 0 {
			c.MsgHandler.SendMsgToTaskQueue(request)
		} else {
			go func(request *Request) {
				c.MsgHandler.DoMsgHandler(request)
			}(request)
		}
	}
}

func (c *Connection) StartWriter() { //专门负责写
	fmt.Println("Writer Goroutine is running", c.ConnID)
	defer fmt.Println("Writer goroutine is stopped", c.ConnID)

	//阻塞等待消息的到来

	for {
		select {
		case msg, ok := <-c.msgChan:
			if !ok {
				fmt.Println("msgChan 已关闭，退出循环")
				return
			}
			//写给客户端
			if _, err := c.Conn.Write(msg); err != nil {
				fmt.Println("write msg error:", err)
				break
			}
		case <-c.ExitChan: //接收到退出信号
			return
		}
	}
}

func (c *Connection) Star() {
	//开始读业务
	go c.StarReader()
	//开启写业务
	go c.StartWriter()
	//TODD 处理业务
	c.server.CallConnStartHook(c)
}

func (c *Connection) Stop() {
	fmt.Println("当前连接关闭,连接id:", c.ConnID, "连接地址:", c.GetRemoteAddr().String())
	if c.isClosed {
		return
	}
	c.isClosed = true
	//告知writer关闭
	c.ExitChan <- true
	//将此链接从链接管理器里删除
	c.server.GetConnManager().RemoveConnection(c)
	c.server.CallConnStopHook(c)
	c.Conn.Close()
	close(c.ExitChan)
	close(c.msgChan)
}

// 产品部 李金雯 研发部 景林
func (c *Connection) SendMessage(msgId uint32, data []byte) error {
	//回写数据给写管理者
	//封包
	message := &Message{
		Id:     msgId,
		Length: uint32(len(data)),
		Msg:    data,
	}
	pack, err := c.Pack.Pack(message)
	if err != nil {
		fmt.Println("Pack error:", err)
		return err
	}
	//将数据发送给管道
	c.msgChan <- pack
	return nil
}
func (c *Connection) GetTCPConnection() net.Conn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) GetRemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(data []byte) error {
	fmt.Println(data)
	return nil
}

func (c *Connection) GetProperty(name string) interface{} { //获取属性
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[name]; ok {
		return value
	}
	return nil
}

func (c *Connection) SetProperty(name string, value interface{}) error { //添加属性
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		fmt.Println("添加属性成功")
		c.property[name] = value
		return nil
	} else {
		fmt.Println("添加属性失败")
		return errors.New("属性名已存在")
	}
}

func (c *Connection) RemoveProperty(name string) { //删除属性
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, name)
}

func (c *Connection) GetServer() ziface.IServer {
	return c.server
}
