package ziface

type IMsgHandle interface {
	DoMsgHandler(request IRequest)               //执行每个消息id对应的处理方法
	AddMsgHandler(uint33 uint32, router IRouter) //添加对应的处理方法
	StartWorkerPoll()                            //开启工作池
	SendMsgToTaskQueue(request IRequest)
	GetWorkPoolSize() uint32
}
