package handler

import (
	"github.com/panjf2000/gnet/v2"
)

// IHandler 定义业务处理接口
type IHandler interface {
	Handle(c gnet.Conn, cmdID uint16, body []byte)
}

// HandlerFuncAdapter 允许普通函数适配 IHandler 接口
type HandlerFuncAdapter func(c gnet.Conn, cmdID uint16, body []byte)

func (h HandlerFuncAdapter) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	h(c, cmdID, body)
}

type Router struct {
	handlers map[uint16]IHandler
}

// NewRouter 创建新的路由实例
func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint16]IHandler),
	}
}

// Register 注册业务处理器
func (r *Router) Register(cmdID uint16, h IHandler) {
	r.handlers[cmdID] = h
}

// RegisterFunc 注册普通函数作为处理器
func (r *Router) RegisterFunc(cmdID uint16, h func(c gnet.Conn, cmdID uint16, body []byte)) {
	r.handlers[cmdID] = HandlerFuncAdapter(h)
}

// Execute 执行对应的处理器
func (r *Router) Execute(c gnet.Conn, cmdID uint16, body []byte) {
	if h, exists := r.handlers[cmdID]; exists {
		h.Handle(c, cmdID, body)
	}

	// 可以在这里加上内存池，回收 body 内存
	// 可以加一个 else 记录未定义路由的日志
}
