package handler

import (
	"github.com/panjf2000/gnet/v2"
)

// HandlerFunc 定义业务处理函数原型
type HandlerFunc func(c gnet.Conn, cmdID uint16, body []byte)

type Router struct {
	handlers map[uint16]HandlerFunc
}

// GlobalRouter 全局单例路由
var GlobalRouter = &Router{
	handlers: make(map[uint16]HandlerFunc),
}

// Register 注册业务处理器
func (r *Router) Register(cmdID uint16, h HandlerFunc) {
	r.handlers[cmdID] = h
}

// Execute 执行对应的处理器
func (r *Router) Execute(c gnet.Conn, cmdID uint16, body []byte) {
	if h, exists := r.handlers[cmdID]; exists {
		h(c, cmdID, body)
	}
	// 可以加一个 else 记录未定义路由的日志
}
