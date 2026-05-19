package network

import (
	"context"
	"runtime"
	"time"

	"gnet_test1/config"
	"gnet_test1/internal/manager"
	"gnet_test1/internal/pool"
	"gnet_test1/internal/protocol"
	"gnet_test1/pkg/logger"

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
	logger.Info().
		Str("version", cfg.App.Version).
		Str("addr", cfg.Server.Addr).
		Msg("🚀 服务器启动成功")
	logger.Info().
		Str("env", cfg.App.Env).
		Str("version", cfg.App.Version).
		Bool("multicore", cfg.Server.Multicore).
		Int("worker_pool_size", cfg.Server.WorkerPoolSize).
		Int("task_queue_size", cfg.Server.TaskQueueSize).
		Int("max_packet_size", cfg.Server.MaxPacketSize).
		Int("heartbeat_check", cfg.Server.HeartbeatCheck).
		Int("heartbeat_timeout", cfg.Server.HeartbeatTimeout).
		Str("log_level", cfg.Log.Level).
		Str("log_path", cfg.Log.Path).
		Bool("log_stdout", cfg.Log.Stdout).
		Msg("📋 配置信息")
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
		logger.Info().
			Int("connections", gs.connMgr.Count()).
			Int("goroutines", runtime.NumGoroutine()).
			Float64("memory_mb", float64(m.Alloc)/1024/1024).
			Msg("📊 状态监控")
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
		logger.Warn().Msg("⚠️ [核心管理] 收到管理员远程关服指令！准备停止全网服务...")
		return gnet.Shutdown
	}

	var connID uint64
	if ctx := c.Context(); ctx != nil {
		if conn, ok := ctx.(manager.Conn); ok {
			connID = conn.ID()
		}
	}

	task := pool.GetWorkTask()
	task.ConnID = connID
	task.CmdID = uint16(cmdID)
	task.Body = payload
	task.Conn = c
	gs.workerPool.Submit(task)

	return gnet.None
}
