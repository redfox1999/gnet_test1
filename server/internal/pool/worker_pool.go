package pool

import (
	"gnet_test1/internal/handler"
	"log"

	"github.com/panjf2000/gnet/v2"
)

// WorkTask 代表投递到 Worker 队列的工作任务
type WorkTask struct {
	ConnID uint64    // 哪个连接发来的
	CmdID  uint16    // 业务指令 ID
	Body   []byte    // 纯业务载荷（已经是深拷贝后的干净数据）
	Conn   gnet.Conn // 维持网络句柄，方便业务层直接回包
}

// Worker 内部独立打工仔
type Worker struct {
	id     int
	ch     chan *WorkTask  // 每个 Worker 独立拥有的任务管道
	router *handler.Router // 绑定的路由处理器
}

func newWorker(id int, queueSize int, router *handler.Router) *Worker {
	return &Worker{
		id:     id,
		ch:     make(chan *WorkTask, queueSize),
		router: router,
	}
}

// start 让 Worker 启动一个常驻 goroutine 监听自己的 channel
func (w *Worker) start() {
	go w.workLoop()
}

func (w *Worker) workLoop() {
	for task := range w.ch {
		if task == nil {
			break
		}
		w.router.Execute(task.Conn, task.CmdID, task.Body)
	}
}

// WorkerPool 统一管理层
type WorkerPool struct {
	workers []*Worker
	size    int
}

// InitWorkerPool 初始化全局固定大小的 Worker 线程池
func InitWorkerPool(poolSize int, queueSize int, r *handler.Router) *WorkerPool {
	pool := &WorkerPool{
		workers: make([]*Worker, poolSize),
		size:    poolSize,
	}
	for i := 0; i < poolSize; i++ {
		pool.workers[i] = newWorker(i, queueSize, r)
		pool.workers[i].start()
	}
	return pool
}

// Submit 根据 ConnID 哈希计算，投递到固定的 Worker 队列里（带高并发防卡死保护）
func (wp *WorkerPool) Submit(task *WorkTask) {
	workerIdx := int(task.ConnID % uint64(wp.size))
	targetWorker := wp.workers[workerIdx]

	// 使用 select 块实现「非阻塞」投递
	select {
	case targetWorker.ch <- task:
		// 队列没满，成功塞入通道，正常安全返回

	default:
		// 🚨 触发保护机制：该 Worker 的业务队列已经爆满了（说明业务处理速度赶不上客户端发包速度）
		log.Printf("[警告] Worker-%d 队列爆满! ConnID: %d 的包被丢弃，可能存在恶意刷包或业务卡顿",
			targetWorker.id, task.ConnID)

		// 生产环境常用做法：
		// 方案 A：直接丢弃该数据包，什么都不做。
		// 方案 B：通知网络层，强制把这个把队伍堵死的客户端连接掐断 (task.Conn.Close())。
		_ = task.Conn.Close()
	}
}
