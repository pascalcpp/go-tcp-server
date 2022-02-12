package znet

import (
	"fmt"
	"go-tcp-server/utils"
	"go-tcp-server/ziface"
	"strconv"
)

type MsgHandler struct {
	Apis           map[uint32]ziface.IRouter
	WorkerPoolSize uint32
	TaskQueues     []chan ziface.IRequest
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueues:     make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

func (mh *MsgHandler) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId =", request.GetMsgID(), "is not found!")
		return
	}

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

func (mh *MsgHandler) AddRouter(msgId uint32, router ziface.IRouter) {
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.FormatUint(uint64(msgId), 10))
	}

	mh.Apis[msgId] = router
	fmt.Println("Add api msgId =", msgId)
}

func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {

	for {
		select {
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

func (mh *MsgHandler) StartWorkerPool() {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		mh.TaskQueues[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		go mh.StartOneWorker(i, mh.TaskQueues[i])
		fmt.Println("Worker ID =", i, " is started.")
	}
}

func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {

	workerID := request.GetConnection().GetConnId() % mh.WorkerPoolSize

	fmt.Println("Add ConnID=", request.GetConnection().GetConnId(),
		" request msgID=", request.GetMsgID(), "to workerID=", workerID)

	mh.TaskQueues[workerID] <- request
}
