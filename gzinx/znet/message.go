package znet

type Message struct {
	Id     uint32
	Msg    []byte
	Length uint32
}

// 获取消息id
func (msg *Message) GetMsgId() uint32 {
	return msg.Id
}

// 获取消息长度
func (msg *Message) GetMsgLen() uint32 {
	return msg.Length
}

// 获取消息内容
func (msg *Message) GetMsg() []byte {
	return msg.Msg
}

// 设置消息ID
func (s *Message) SetMsgId(id uint32) {
	s.Id = id
}

// 设置消息内容
func (s *Message) SetMsg(data []byte) {
	s.Msg = data
}

// 设置消息长度
func (s *Message) SetMsgLen(length uint32) {
	s.Length = length
}
