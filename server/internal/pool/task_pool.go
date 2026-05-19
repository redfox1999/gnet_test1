package pool

import (
	"sync"

	"github.com/panjf2000/gnet/v2"
)

// WorkTask 代表投递到 Worker 队列的工作任务
type WorkTask struct {
	ConnID  uint64    // 哪个连接发来的
	CmdID   uint32    // 业务指令 ID
	Body    []byte    // 纯业务载荷（已经是深拷贝后的干净数据）
	DataLen int       // 数据长度
	Conn    gnet.Conn // 维持网络句柄，方便业务层直接回包
}

var WorkTaskPool = sync.Pool{
	New: func() interface{} {
		return &WorkTask{}
	},
}

func GetWorkTask() *WorkTask {
	return WorkTaskPool.Get().(*WorkTask)
}

func PutWorkTask(task *WorkTask) {
	if task.Body != nil {
		PutBytes(task.Body)
	}
	task.ConnID = 0
	task.CmdID = 0
	task.DataLen = 0
	task.Body = nil
	task.Conn = nil
	WorkTaskPool.Put(task)
}
