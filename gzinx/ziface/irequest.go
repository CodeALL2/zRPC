package ziface

type IRequest interface {
	GetConnection() IConnection
	GetData() []byte
	GetMsgID() uint32
	GetMsgLen() uint32
}
