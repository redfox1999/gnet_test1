package network

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"gnet_test1/config"
	"gnet_test1/internal/pool"

	"github.com/panjf2000/gnet/v2"
)

type UserContext struct {
	ConnID     int64
	UID        int64
	Username   string
	IsLoggedIn bool
	LoginTime  time.Time
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
	log.Printf("🚀 工业级网关已成功在 %s 启动！监听中...", gs.cfg.Addr)
	gs.engine = eng // 保存引擎句柄

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
		ConnID:     connID,
		IsLoggedIn: false,
		LoginTime:  time.Now(),
	}
	c.SetContext(ctx)
	atomic.AddInt64(&gs.connCount, 1)
	log.Printf("[连接成功] ConnID=%d，来自客户端: %s，当前在线: %d", connID,
		c.RemoteAddr().String(), atomic.LoadInt64(&gs.connCount))
	return nil, gnet.None
}

func (gs *GatewayServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	atomic.AddInt64(&gs.connCount, -1)
	if ctx, ok := c.Context().(*UserContext); ok && ctx.IsLoggedIn {
		log.Printf("[连接断开] ConnID=%d，用户离线: UID=%d, 名字=%s，当前在线: %d，错误=%v",
			ctx.ConnID, ctx.UID, ctx.Username, atomic.LoadInt64(&gs.connCount), err)
	} else {
		var connID int64
		if ctx, ok := c.Context().(*UserContext); ok {
			connID = ctx.ConnID
		}
		log.Printf("[连接断开] ConnID=%d，未登录客户端断开: %s，当前在线: %d", connID,
			c.RemoteAddr().String(), atomic.LoadInt64(&gs.connCount))
	}
	return gnet.None
}

func (gs *GatewayServer) dispatchBusiness(c gnet.Conn, cmdID uint16, seq uint32, payload []byte) gnet.Action {
	ctx, _ := c.Context().(*UserContext)

	if cmdID == CmdShutDown {
		log.Println("⚠️ [核心管理] 收到管理员远程关服指令！准备停止全网服务...")
		return gnet.Shutdown
	}

	// 构建工作任务并提交到线程池
	task := &pool.WorkTask{
		ConnID: uint64(ctx.ConnID),
		CmdID:  cmdID,
		Body:   payload,
		Conn:   c,
	}
	gs.workerPool.Submit(task)

	return gnet.None
}
