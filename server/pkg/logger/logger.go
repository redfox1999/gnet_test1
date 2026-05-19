package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/rs/zerolog"
)

var (
	log          zerolog.Logger
	consoleLog   zerolog.Logger
	initialized  bool
	gnetLogLevel zerolog.Level
)

// Config 日志配置
type Config struct {
	Level      string // 业务日志级别：debug, info, warn, error
	Path       string // 日志文件路径
	Stdout     bool   // 是否输出到标准输出
	Filename   string // 日志文件名
	MaxSize    int64  // 最大文件大小 (MB)
	MaxBackups int    // 最大备份文件数
	MaxAge     int    // 最大保存天数
	GnetLevel  string // gnet 日志级别：debug, info, warn, error（与业务层分开）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Path:       "./logs",
		Stdout:     true,
		Filename:   "server.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     30,
		GnetLevel:  "warn", // gnet 默认级别设为 warn，避免过多日志
	}
}

// Init 初始化日志系统
func Init(cfg *Config) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 设置 gnet 日志级别
	gnetLevel, err := zerolog.ParseLevel(cfg.GnetLevel)
	if err != nil {
		gnetLevel = zerolog.WarnLevel
	}
	gnetLogLevel = gnetLevel

	// 设置业务日志级别
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// 设置输出
	var fileWriters []io.Writer
	var consoleWriters []io.Writer

	// 如果启用标准输出，添加到控制台输出
	if cfg.Stdout {
		consoleWriters = append(consoleWriters, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05.000",
		})
	}

	// 如果配置了日志文件路径，创建文件输出
	if cfg.Path != "" {
		if err := os.MkdirAll(cfg.Path, 0755); err != nil {
			// 创建目录失败，只输出到 stdout
			log = zerolog.New(zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: "2006-01-02 15:04:05.000",
			}).With().Timestamp().Caller().Logger()
			consoleLog = log
			initialized = true
			return
		}

		logFile := filepath.Join(cfg.Path, cfg.Filename)
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fileWriters = append(fileWriters, file)
		}
	}

	// 创建业务日志（文件 + 控制台）
	var combinedWriters []io.Writer
	combinedWriters = append(combinedWriters, fileWriters...)
	combinedWriters = append(combinedWriters, consoleWriters...)

	if len(combinedWriters) > 0 {
		writer := io.MultiWriter(combinedWriters...)
		log = zerolog.New(writer).With().Timestamp().Caller().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		log = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	}

	// 创建控制台日志（仅用于确保彩色输出）
	if len(consoleWriters) > 0 {
		consoleWriter := io.MultiWriter(consoleWriters...)
		consoleLog = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
	} else {
		consoleLog = log
	}

	initialized = true
}

// Get 获取全局 logger 实例
func Get() *zerolog.Logger {
	if !initialized {
		Init(nil)
	}
	return &log
}

// GetConsole 获取控制台 logger（带彩色输出）
func GetConsole() *zerolog.Logger {
	if !initialized {
		Init(nil)
	}
	return &consoleLog
}

// 基础日志方法

// Debug 开始一条 debug 级别日志
func Debug() *zerolog.Event {
	return Get().Debug()
}

// Info 开始一条 info 级别日志
func Info() *zerolog.Event {
	return Get().Info()
}

// Warn 开始一条 warn 级别日志
func Warn() *zerolog.Event {
	return Get().Warn()
}

// Error 开始一条 error 级别日志
func Error() *zerolog.Event {
	return Get().Error()
}

// Fatal 开始一条 fatal 级别日志
func Fatal() *zerolog.Event {
	return Get().Fatal()
}

// Panic 开始一条 panic 级别日志
func Panic() *zerolog.Event {
	return Get().Panic()
}

// With 添加上下文
func With() zerolog.Context {
	return Get().With()
}

// 便捷日志方法

// Debugf 格式化输出 debug 日志
func Debugf(format string, v ...interface{}) {
	Get().Debug().Msgf(format, v...)
}

// Infof 格式化输出 info 日志
func Infof(format string, v ...interface{}) {
	Get().Info().Msgf(format, v...)
}

// Warnf 格式化输出 warn 日志
func Warnf(format string, v ...interface{}) {
	Get().Warn().Msgf(format, v...)
}

// Errorf 格式化输出 error 日志
func Errorf(format string, v ...interface{}) {
	Get().Error().Msgf(format, v...)
}

// 业务场景日志方法

// ServerStarted 记录服务器启动
func ServerStarted(addr, version, gitCommit string) {
	Get().Info().
		Str("version", version).
		Str("git_commit", gitCommit).
		Str("addr", addr).
		Msg("🚀 服务器启动成功")
}

// ServerShutdown 记录服务器关闭
func ServerShutdown(signal string, duration time.Duration) {
	Get().Info().
		Str("signal", signal).
		Dur("duration", duration).
		Msg("⚠️ 服务器已关闭")
}

// ConnOpened 记录连接建立
func ConnOpened(connID uint64, addr string) {
	Get().Info().
		Uint64("conn_id", connID).
		Str("remote_addr", addr).
		Msg("📡 新连接建立")
}

// ConnClosed 记录连接关闭
func ConnClosed(connID uint64, addr string, err error, duration time.Duration) {
	event := Get().Info().
		Uint64("conn_id", connID).
		Str("remote_addr", addr).
		Dur("duration", duration)
	if err != nil {
		event.Err(err)
	}
	event.Msg("🔌 连接已关闭")
}

// BusinessProcessed 记录业务处理完成
func BusinessProcessed(connID uint64, cmdID uint16, duration time.Duration) {
	Get().Debug().
		Uint64("conn_id", connID).
		Uint16("cmd_id", cmdID).
		Dur("duration", duration).
		Msg("⚙️ 业务处理完成")
}

// WorkerQueueFull 记录工作队列已满
func WorkerQueueFull(workerID int, connID uint64) {
	Get().Warn().
		Int("worker_id", workerID).
		Uint64("conn_id", connID).
		Msg("⚠️ Worker 队列已满，连接已关闭")
}

// StatsReport 定期状态报告
func StatsReport(connCount int, goroutineCount int, memoryMB float64) {
	Get().Info().
		Int("connections", connCount).
		Int("goroutines", goroutineCount).
		Float64("memory_mb", memoryMB).
		Msg("📊 状态监控")
}

// GnetLoggerAdapter 适配 gnet Logger 接口的适配器（使用独立的日志级别）
type GnetLoggerAdapter struct{}

// Debugf 实现 gnet Logger 接口（根据 gnet 级别过滤）
func (g *GnetLoggerAdapter) Debugf(format string, args ...any) {
	if gnetLogLevel <= zerolog.DebugLevel {
		GetConsole().Debug().Msgf(format, args...)
	}
}

// Infof 实现 gnet Logger 接口（根据 gnet 级别过滤）
func (g *GnetLoggerAdapter) Infof(format string, args ...any) {
	if gnetLogLevel <= zerolog.InfoLevel {
		GetConsole().Info().Msgf(format, args...)
	}
}

// Warnf 实现 gnet Logger 接口（根据 gnet 级别过滤）
func (g *GnetLoggerAdapter) Warnf(format string, args ...any) {
	if gnetLogLevel <= zerolog.WarnLevel {
		GetConsole().Warn().Msgf(format, args...)
	}
}

// Errorf 实现 gnet Logger 接口（根据 gnet 级别过滤）
func (g *GnetLoggerAdapter) Errorf(format string, args ...any) {
	if gnetLogLevel <= zerolog.ErrorLevel {
		GetConsole().Error().Msgf(format, args...)
	}
}

// Fatalf 实现 gnet Logger 接口
func (g *GnetLoggerAdapter) Fatalf(format string, args ...any) {
	GetConsole().Fatal().Msgf(format, args...)
}

// GnetFlusher 返回 gnet 的 Flusher 函数
func GnetFlusher() logging.Flusher {
	return func() error {
		return nil
	}
}

// SetGnetDefaultLoggerAndFlusher 将自定义日志设置为 gnet 的默认日志
func SetGnetDefaultLoggerAndFlusher() {
	logging.SetDefaultLoggerAndFlusher(&GnetLoggerAdapter{}, GnetFlusher())
}
