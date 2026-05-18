package network

import (
	"context"
	"log"
	"sync/atomic"

	"gnet_test1/config"
	"gnet_test1/internal/pool"
	"gnet_test1/internal/protocol"

	"github.com/panjf2000/gnet/v2"
)

type UserContext struct {
	ConnID int64
}

type GatewayServer struct {
	gnet.BuiltinEventEngine
	cfg        *config.ServerConfig
	connCount  int64
	nextConnID int64
	engine     gnet.Engine      // 用于暂存 gnet 引擎句柄
	workerPool *pool.WorkerPool // 业务线程池
}

func NewGatewayServer(cfg *config.ServerConfig, workerPool *pool.WorkerPool) *GatewayServer {
	return &GatewayServer{
		cfg:        cfg,
		workerPool: workerPool,
	}
}

func (gs *GatewayServer) Addr() string {
	return gs.cfg.Addr
}

func (gs *GatewayServer) OnBoot(eng gnet.Engine) gnet.Action {
	cfg := config.Global
	log.Printf("🚀 工业级网关已成功在 %s 启动！", cfg.Server.Addr)
	log.Printf("📋 配置信息:")
	log.Printf("   App: env=%s, version=%s", cfg.App.Env, cfg.App.Version)
	log.Printf("   Server: multicore=%v, worker_pool_size=%d, task_queue_size=%d, max_packet_size=%d",
		cfg.Server.Multicore, cfg.Server.WorkerPoolSize, cfg.Server.TaskQueueSize, cfg.Server.MaxPacketSize)
	log.Printf("   Server: heartbeat_check=%ds, heartbeat_timeout=%ds", cfg.Server.HeartbeatCheck, cfg.Server.HeartbeatTimeout)
	log.Printf("   Log: level=%s, path=%s, stdout=%v", cfg.Log.Level, cfg.Log.Path, cfg.Log.Stdout)
	gs.engine = eng

	return gnet.None
}

// CloseEngine 关闭引擎
func (gs *GatewayServer) CloseEngine() {
	gs.engine.Stop(context.Background())
}

// Start 启动服务器
func (gs *GatewayServer) Start() error {
	return gnet.Run(gs, gs.cfg.Addr, gnet.WithMulticore(gs.cfg.Multicore))
}

func (gs *GatewayServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	connID := atomic.AddInt64(&gs.nextConnID, 1)
	ctx := &UserContext{
		ConnID: connID,
	}
	c.SetContext(ctx)
	atomic.AddInt64(&gs.connCount, 1)
	//log.Printf("[连接成功] ConnID=%d，来自客户端: %s，当前在线: %d", connID,
	//	c.RemoteAddr().String(), atomic.LoadInt64(&gs.connCount))
	return nil, gnet.None
}

func (gs *GatewayServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	atomic.AddInt64(&gs.connCount, -1)
	//var connID int64
	// if ctx, ok := c.Context().(*UserContext); ok {
	// 	connID = ctx.ConnID
	// }
	//log.Printf("[连接断开] ConnID=%d，客户端断开: %s，当前在线: %d，错误=%v",
	//	connID, c.RemoteAddr().String(), atomic.LoadInt64(&gs.connCount), err)
	return gnet.None
}

func (gs *GatewayServer) dispatchBusiness(c gnet.Conn, cmdID uint32, payload []byte) gnet.Action {
	ctx, _ := c.Context().(*UserContext)

	if cmdID == protocol.CmdShutDown {
		log.Println("⚠️ [核心管理] 收到管理员远程关服指令！准备停止全网服务...")
		return gnet.Shutdown
	}

	// 构建工作任务并提交到线程池
	task := &pool.WorkTask{
		ConnID: uint64(ctx.ConnID),
		CmdID:  uint16(cmdID),
		Body:   payload,
		Conn:   c,
	}
	gs.workerPool.Submit(task)

	return gnet.None
}
