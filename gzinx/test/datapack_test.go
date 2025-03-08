package test

import (
	"fmt"
	"go-tcp-server/gzinx/znet"
	"net"
	"testing"
	"time"
)

// 测试封包拆包的单元测试
func TestDataPack(t *testing.T) {
	//
	////创建一个tcp的服务
	//
	//listen, err := net.Listen("tcp", ":9888")
	//
	//if err != nil {
	//	return
	//}
	//defer listen.Close()
	//
	//go func() {
	//	for {
	//		conn, err := listen.Accept()
	//		if err != nil {
	//			return
	//		}
	//		//go run开启客户端的链接
	//		go func(conn net.Conn) {
	//			defer conn.Close()
	//			//开始读取包
	//			packUtil := znet.NewDataPack()
	//			//读取head
	//			for {
	//
	//				headLen := make([]byte, packUtil.GetPackHeadLen())
	//				//读取head
	//				if _, err := io.ReadFull(conn, headLen); err != nil {
	//					fmt.Println(err, "读取head失败")
	//					return
	//				}
	//				//拆包
	//				messageHead, err2 := packUtil.Unpack(headLen)
	//				if err2 != nil {
	//					fmt.Println(err2)
	//					return
	//				}
	//				if messageHead.GetMsgLen() > 0 {
	//					body := make([]byte, messageHead.GetMsgLen())
	//					//读取包的内容
	//					_, err2 = io.ReadFull(conn, body)
	//					if err2 != nil {
	//						fmt.Println(err2)
	//						return
	//					}
	//					messageHead.SetMsg(body)
	//				}
	//
	//				fmt.Println("Msg ID", messageHead.GetMsgId(), "Msg Len", messageHead.GetMsgLen(), "Msg body", messageHead.GetMsg())
	//			}
	//
	//		}(conn)
	//
	//	}
	//}()

	dial, err := net.Dial("tcp", "127.0.0.1:9888")
	if err != nil {
		return
	}
	defer dial.Close()
	fmt.Println("客户端已经连接")

	//开始封包
	message1 := &znet.Message{
		Id:     1,
		Length: 4,
		Msg:    []byte{'a', 'b', 'c', 'd'},
	}

	pack := znet.NewDataPack()
	bytes1, err := pack.Pack(message1)

	//开始封包
	message2 := &znet.Message{
		Id:     2,
		Length: 4,
		Msg:    []byte{'z', 'i'},
	}

	bytes2, err := pack.Pack(message2)
	bytes := append(bytes1, bytes2...)

	dial.Write(bytes)
	time.Sleep(2 * time.Second)
	dial.Write([]byte{'h', 'h'})

	select {
	case <-time.After(15 * time.Second):
		fmt.Println("测试结束")
	}

}
