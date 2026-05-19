package network

import (
	"context"
	"log"
	"runtime"
	"time"

	"gnet_test1/config"
	"gnet_test1/internal/manager"
	"gnet_test1/internal/pool"
	"gnet_test1/internal/protocol"

	"github.com/panjf2000/gnet/v2"
)

type GatewayServer struct {
	gnet.BuiltinEventEngine
	cfg        *config.ServerConfig
	engine     gnet.Engine          // 用于暂存 gnet 引擎句柄
	workerPool *pool.WorkerPool     // 业务线程池
	connMgr    *manager.ConnManager // 连接管理器
}

func NewGatewayServer(cfg *config.ServerConfig, workerPool *pool.WorkerPool) *GatewayServer {
	return &GatewayServer{
		cfg:        cfg,
		workerPool: workerPool,
		connMgr:    manager.NewConnManager(),
	}
}

func (gs *GatewayServer) Addr() string {
	return gs.cfg.Addr
}

func (gs *GatewayServer) OnBoot(eng gnet.Engine) gnet.Action {
	cfg := config.Global
	log.Printf("🚀 网关已成功在 %s 启动！", cfg.Server.Addr)
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
	go gs.statsReporter()
	return gnet.Run(gs, gs.cfg.Addr, gnet.WithMulticore(gs.cfg.Multicore))
}

func (gs *GatewayServer) statsReporter() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var m runtime.MemStats
	for {
		<-ticker.C
		runtime.ReadMemStats(&m)
		log.Printf("📊 状态监控: 连接数=%d, Goroutine数=%d, 内存占用=%.2fMB",
			gs.connMgr.Count(),
			runtime.NumGoroutine(),
			float64(m.Alloc)/1024/1024)
	}
}

func (gs *GatewayServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	connID := gs.connMgr.NextID()
	conn := manager.NewGnetConn(c, connID)
	gs.connMgr.Add(conn)
	return nil, gnet.None
}

func (gs *GatewayServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if ctx := c.Context(); ctx != nil {
		if conn, ok := ctx.(manager.Conn); ok {
			gs.connMgr.Remove(conn.ID())
		}
	}
	return gnet.None
}

func (gs *GatewayServer) dispatchBusiness(c gnet.Conn, cmdID uint32, payload []byte) gnet.Action {
	if cmdID == protocol.CmdShutDown {
		log.Println("⚠️ [核心管理] 收到管理员远程关服指令！准备停止全网服务...")
		return gnet.Shutdown
	}

	var connID uint64
	if ctx := c.Context(); ctx != nil {
		if conn, ok := ctx.(manager.Conn); ok {
			connID = conn.ID()
		}
	}

	task := &pool.WorkTask{
		ConnID: connID,
		CmdID:  uint16(cmdID),
		Body:   payload,
		Conn:   c,
	}
	gs.workerPool.Submit(task)

	return gnet.None
}
