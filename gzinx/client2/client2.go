package main

import (
	"fmt"
	"go-tcp-server/gzinx/znet"
	"io"
	"net"
	"time"
)

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:9888")
	if err != nil {
		fmt.Println("连接失败", err)
	}
	defer conn.Close()
	bytes := []byte("2  hello zinx server")
	//封装一个数据报文
	message := &znet.Message{
		Id:     2,
		Length: uint32(len(bytes)),
		Msg:    bytes,
	}

	packUtils := znet.NewDataPack()
	//封包
	pack, err := packUtils.Pack(message)
	if err != nil {
		return
	}

	for {
		_, err := conn.Write(pack)
		if err != nil {
			return
		}
		head := make([]byte, 8)
		if _, err := io.ReadFull(conn, head); err != nil {
			return
		}

		//读取完头信息后, 将消息拆包
		message, err := packUtils.Unpack(head)
		if err != nil {
			return
		}

		body := make([]byte, message.GetMsgLen())
		if _, err = io.ReadFull(conn, body); err != nil {
			return
		}
		message.SetMsg(body)
		//打印出服务器传回来的消息
		fmt.Println("msgId:", message.GetMsgId(), "msgLen:", message.GetMsgLen(), "msgBody", string(message.GetMsg()))
		time.Sleep(2 * time.Second)
	}

}
