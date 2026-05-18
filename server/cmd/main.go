package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gnet_test1/config"
	"gnet_test1/internal/handler"
	"gnet_test1/internal/network"
	"gnet_test1/internal/pool"
)

func main() {
	// 解析命令行参数
	cfgPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置文件
	_, err := config.InitConfig(*cfgPath)
	if err != nil {
		log.Printf("[警告] 加载配置文件失败: %v，将使用默认配置", err)
	}

	// 初始化业务协程池
	workerPool := pool.InitWorkerPool(config.Global.Server.WorkerPoolSize,
		config.Global.Server.TaskQueueSize,
		handler.GlobalRouter)
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

		// 3. 阻塞等待关服信号到来（刚才报错是因为上面漏了第1步，导致这里 undefined）
		sig := <-sigCh
		// 监听键盘 Ctrl+C (Interrupt) 和 Linux kill 信号 (Terminate)
		log.Printf("⚠️ 捕获到系统退出信号 [%v]，正在启动优雅关闭流程...", sig)

		// 阻塞等待关服信号到来
		server.CloseEngine()
		log.Println("🎉 服务器优雅退出成功！")
		os.Exit(0)
		// 第一步：通知 gnet 引擎关闭。
		// 你的 server 结构体需要暴露出内部的 gnet.Engine 实例（可以通过 OnBoot 事件拿到）
		// 调用 engine.Stop() 会停止接收新连接，并让 main 里的 gnet.Run 优雅解除阻塞
	}()

	server.Start()

	// 第二步：关闭业务线程池，确保正在处理的数据不丢失
	//pool.GlobalPool.Shutdown()

	log.Print("服务已停止")

}
