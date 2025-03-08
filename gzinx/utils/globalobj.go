package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"zRPC/gzinx/ziface"
)

type GlobalObj struct {
	/*
		Server
	*/
	TcpServer ziface.IServer //当前的全局对象
	Host      string
	Port      int32
	Name      string

	/*
		Gzinx
	*/
	Version       string //当前的版本号
	MaxConn       int32  //最大连接数
	MaxPacketSize uint32 //接收数据包的最大容量

	MaxWorkPoolSize uint32 //工作池的大小
	MaxTaskQueueLen uint32 //队列的长度
}

var GlobalObject *GlobalObj

func Reload(path string) {
	// 获取当前工作目录
	fmt.Println(path)
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

func init() {
	GlobalObject = &GlobalObj{
		Name:            "zRPC",
		Host:            "0.0.0.0",
		Port:            9888,
		Version:         "V0.1",
		MaxConn:         500,
		MaxPacketSize:   4096,
		MaxWorkPoolSize: 10,
		MaxTaskQueueLen: 1000,
	}
}
