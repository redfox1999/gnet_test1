package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// GlobalConfig 全局配置单例
var (
	Global *Config
	once   sync.Once
)

// Default 默认配置常量
var Default = &Config{
	App: AppConfig{
		Env:     "dev",
		Version: "1.0.0",
	},
	Server: ServerConfig{
		Addr:             "tcp://0.0.0.0:9000",
		Multicore:        true,
		WorkerPoolSize:   10,   // 业务协程池（Goroutine Pool）的大小
		TaskQueueSize:    1024, // 每个 Worker 协程的任务队列大小
		MaxPacketSize:    65535,
		HeartbeatCheck:   30,
		HeartbeatTimeout: 90,
	},
	Log: LogConfig{
		Level:  "info",
		Path:   "./logs/",
		Stdout: true,
	},
}

// Config 对应整个配置文件结构的根节点
type Config struct {
	App    AppConfig    `yaml:"app"`
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Env     string `yaml:"env"`     // 运行环境: dev, test, prod
	Version string `yaml:"version"` // 系统版本号
}

// ServerConfig 网络层与线程池深度参数（针对 gnet 调优）
type ServerConfig struct {
	Addr             string `yaml:"addr"`              // 监听地址，例如: tcp://127.0.0.1:9000
	Multicore        bool   `yaml:"multicore"`         // 是否开启 gnet 的多核心绑定
	WorkerPoolSize   int    `yaml:"worker_pool_size"`  // 业务线程池（Goroutine Pool）的大小
	TaskQueueSize    int    `yaml:"task_queue_size"`   // 任务队列大小，用于缓冲待处理的任务，避免阻塞
	MaxPacketSize    int    `yaml:"max_packet_size"`   // 单个包最大限制（单位: 字节），用于防御恶意攻击
	HeartbeatCheck   int    `yaml:"heartbeat_check"`   // 心跳检测间隔时间（单位: 秒）
	HeartbeatTimeout int    `yaml:"heartbeat_timeout"` // 判定心跳超时强制断开的时间（单位: 秒）
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // 日志级别: debug, info, warn, error
	Path   string `yaml:"path"`   // 日志输出路径
	Stdout bool   `yaml:"stdout"` // 是否同时输出到控制台
}

// InitConfig 全局初始化加载函数（单例模式，保证只加载一次）
func InitConfig(filePath string) (*Config, error) {
	var err error
	once.Do(func() {
		Global = &Config{}
		err = Global.load(filePath)
	})
	return Global, err
}

// load 内部私有加载方法
func (c *Config) load(filePath string) error {
	// 先设置默认值，确保即使加载失败也有兜底值
	c.setDefaults()

	// 读取配置文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 解析 YAML，成功后会覆盖默认值
	return yaml.Unmarshal(data, c)
}

// setDefaults 兜底默认值设置，使用 Default 配置
func (c *Config) setDefaults() {
	if c.App.Env == "" {
		c.App.Env = Default.App.Env
	}
	if c.App.Version == "" {
		c.App.Version = Default.App.Version
	}
	if c.Server.Addr == "" {
		c.Server.Addr = Default.Server.Addr
	}
	if !c.Server.Multicore {
		c.Server.Multicore = Default.Server.Multicore
	}
	if c.Server.WorkerPoolSize <= 0 {
		c.Server.WorkerPoolSize = Default.Server.WorkerPoolSize
	}
	if c.Server.MaxPacketSize <= 0 {
		c.Server.MaxPacketSize = Default.Server.MaxPacketSize
	}
	if c.Server.HeartbeatCheck <= 0 {
		c.Server.HeartbeatCheck = Default.Server.HeartbeatCheck
	}
	if c.Server.HeartbeatTimeout <= 0 {
		c.Server.HeartbeatTimeout = Default.Server.HeartbeatTimeout
	}
	if c.Log.Level == "" {
		c.Log.Level = Default.Log.Level
	}
	if c.Log.Path == "" {
		c.Log.Path = Default.Log.Path
	}
}
