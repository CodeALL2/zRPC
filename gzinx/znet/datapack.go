package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zRPC/gzinx/utils"
	"zRPC/gzinx/ziface"
)

type DataPack struct{}

func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包的head
func (p *DataPack) GetPackHeadLen() uint32 {
	//4字节长度 + 4字节的id
	return 8
}

// 封包
func (p *DataPack) Pack(message ziface.IMessage) ([]byte, error) {
	//创建一个字节数组
	buffer := bytes.NewBuffer([]byte{})
	//写入包头
	err := binary.Write(buffer, binary.LittleEndian, message.GetMsgLen())

	if err != nil {
		return nil, err
	}

	//写入包的id
	err = binary.Write(buffer, binary.LittleEndian, message.GetMsgId())

	if err != nil {
		return nil, err
	}

	//写入包的内容
	err = binary.Write(buffer, binary.LittleEndian, message.GetMsg())
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// 拆包方法 --> 拆的是头
func (p *DataPack) Unpack(message []byte) (ziface.IMessage, error) {
	buffer := bytes.NewBuffer(message)
	msg := &Message{}
	err := binary.Read(buffer, binary.LittleEndian, &msg.Length)

	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.LittleEndian, &msg.Id)

	if err != nil {
		return nil, err
	}

	if utils.GlobalObject.MaxPacketSize > 0 && msg.Length > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("too large msg data received")
	}

	return msg, nil
}
