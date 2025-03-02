package ziface

type IDataPack interface {
	GetPackHeadLen() uint32                  //获取头的长度
	Pack(message IMessage) ([]byte, error)   //封包
	Unpack(message []byte) (IMessage, error) //拆包
}
