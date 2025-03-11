package iface

type IClientMsgHandle interface {
	ClientStartWorkerPoll()
	DoMsgHandler(request IClientMsgRequest)
	AddMsgHandler(uint33 uint32, router IClientRouter)
	SendMessageToClientQueue(request IClientMsgRequest)
}
