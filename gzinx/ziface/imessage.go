package ziface

type IMessage interface {
	//获取消息id
	GetMsgId() uint32
	//获取消息长度
	GetMsgLen() uint32
	//获取消息内容
	GetMsg() []byte
	//设置消息ID
	SetMsgId(uint32)
	//设置消息内容
	SetMsg([]byte)
	//设置消息长度
	SetMsgLen(uint32)
}
