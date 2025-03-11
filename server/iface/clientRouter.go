package iface

type IClientRouter interface {
	Handler(msg IClientMsgRequest)
}
