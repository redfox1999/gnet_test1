package handler

import (
	"gnet_test1/internal/manager"

	"github.com/panjf2000/gnet/v2"
)

// IHandler 定义业务处理接口
type IHandler interface {
	Handle(c gnet.Conn, cmdID uint32, body []byte)
}

// HandlerFuncAdapter 允许普通函数适配 IHandler 接口
type HandlerFuncAdapter func(c gnet.Conn, cmdID uint32, body []byte)

func (h HandlerFuncAdapter) Handle(c gnet.Conn, cmdID uint32, body []byte) {
	h(c, cmdID, body)
}

type Router struct {
	handlers map[uint32]IHandler
}

// NewRouter 创建新的路由实例
func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint32]IHandler),
	}
}

// Register 注册业务处理器
func (r *Router) Register(cmdID uint32, h IHandler) {
	r.handlers[cmdID] = h
}

// RegisterFunc 注册普通函数作为处理器
func (r *Router) RegisterFunc(cmdID uint32, h func(c gnet.Conn, cmdID uint32, body []byte)) {
	r.handlers[cmdID] = HandlerFuncAdapter(h)
}

// Execute 执行对应的处理器
func (r *Router) Execute(c gnet.Conn, cmdID uint32, body []byte) {
	if h, exists := r.handlers[cmdID]; exists {
		h.Handle(c, cmdID, body)
		conn := c.Context().(manager.Conn)
		conn.DelPendingTask()
	}

	// 可以加一个 else 记录未定义路由的日志
}
