package znet

import (
	"fmt"
	"zRPC/gzinx/utils"
	"zRPC/gzinx/ziface"
)

type MsgHandle struct {
	MsgHandleMap   map[uint32]ziface.IRouter
	WorkerPoolSize uint32                 //业务工作worker池的数量
	TaskQueue      []chan ziface.IRequest //worker负责取任务的消息队列
}

var index uint32 = 0

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		MsgHandleMap:   make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.MaxWorkPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.MaxWorkPoolSize),
	}
}
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) { //执行每个消息id对应的处理方法

	router, ok := mh.MsgHandleMap[request.GetMsgID()]
	if !ok {
		fmt.Println("do msg handler error :", request.GetMsgID())
		return
	}
	router.PreHandler(request)
	router.Handler(request)
	router.PostHandler(request)
}

func (mh *MsgHandle) AddMsgHandler(uint33 uint32, router ziface.IRouter) { //添加对应的执行方法

	if mh.MsgHandleMap == nil || mh.MsgHandleMap[uint33] != nil {
		fmt.Println("add msg handler error :", uint33)
		return
	}

	mh.MsgHandleMap[uint33] = router
}

func (mh *MsgHandle) StartWorkerPoll() {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxTaskQueueLen)
		go mh.startWorker(i, mh.TaskQueue[i]) //启动一个worker
	}
}

func (mh *MsgHandle) startWorker(taskQueueId int, queue chan ziface.IRequest) {
	fmt.Println("工作池队列", taskQueueId, "已经启动")
	for {
		select {
		case msg := <-queue:
			mh.DoMsgHandler(msg)
		}
	}
}

func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	queue := mh.TaskQueue[index%mh.WorkerPoolSize]
	index++
	queue <- request
}

func (mg *MsgHandle) GetWorkPoolSize() uint32 {
	return mg.WorkerPoolSize
}
