package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gnet_test1/config"
	"gnet_test1/internal/handler"
	"gnet_test1/internal/network"
	"gnet_test1/internal/pool"
	"gnet_test1/internal/protocol"
	"gnet_test1/pkg/logger"
)

var (
	Version    = "unknown"
	GitCommit  = "unknown"
	CommitTime = "unknown"
	BuildTime  = "unknown"
)

func main() {
	// 解析命令行参数
	cfgPath := flag.String("config", "config/config.yaml", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.BoolVar(showVersion, "V", false, "显示版本信息（简写）")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("版本：%s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Commit Time: %s\n", CommitTime)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}

	// 加载配置文件
	_, err := config.InitConfig(*cfgPath)
	if err != nil {
		logger.Warn().Err(err).Msg("加载配置文件失败，将使用默认配置")
	}

	// 初始化日志系统
	logger.Init(&logger.Config{
		Level:      config.Global.Log.Level,
		GnetLevel:  config.Global.Log.GnetLevel,
		Path:       config.Global.Log.Path,
		Stdout:     config.Global.Log.Stdout,
		Filename:   config.Global.Log.Filename,
		MaxSize:    config.Global.Log.MaxSize,
		MaxBackups: config.Global.Log.MaxBackups,
		MaxAge:     config.Global.Log.MaxAge,
	})

	// 将自定义日志设置为 gnet 的默认日志
	logger.SetGnetDefaultLoggerAndFlusher()

	// 创建路由实例
	router := handler.NewRouter()

	// 注册测试用 Handler
	router.Register(uint16(protocol.CmdCalculate), &handler.CalculateHandler{})
	router.Register(uint16(protocol.CmdSmall), &handler.SmallHandler{})
	router.Register(uint16(protocol.CmdMedium), &handler.MediumHandler{})
	router.Register(uint16(protocol.CmdLarge), &handler.LargeHandler{})

	// 初始化业务协程池
	workerPool := pool.InitWorkerPool(config.Global.Server.WorkerPoolSize,
		config.Global.Server.TaskQueueSize, router)
	// 使用配置创建服务器
	server := network.NewGatewayServer(&config.Global.Server, workerPool)

	// 开启独立协程监听系统信号（优雅退出）
	go func() {
		// 3. 💡【核心优雅退出控制】开启一个独立的协程监听系统信号

		// 创建信号接管通道
		// 1. 确保这里使用 := 正确声明并初始化了 sigCh 变量
		sigCh := make(chan os.Signal, 1)

		// 2. 将通道注册到系统的信号监听器中
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// 3. 阻塞等待关服信号到来（刚才报错是因为上面漏了第 1 步，导致这里 undefined）
		sig := <-sigCh
		// 监听键盘 Ctrl+C (Interrupt) 和 Linux kill 信号 (Terminate)
		logger.Info().Str("signal", sig.String()).Msg("⚠️ 服务器已关闭")

		// 阻塞等待关服信号到来
		server.CloseEngine()
		logger.Info().Msg("🎉 服务器优雅退出成功！")
		// 第一步：通知 gnet 引擎关闭。
		// 你的 server 结构体需要暴露出内部的 gnet.Engine 实例（可以通过 OnBoot 事件拿到）
		// 调用 engine.Stop() 会停止接收新连接，并让 main 里的 gnet.Run 优雅解除阻塞
	}()

	server.Start()

	// 第二步：关闭业务线程池，确保正在处理的数据不丢失
	//pool.GlobalPool.Shutdown()

	logger.Info().Msg("服务已停止")

}
