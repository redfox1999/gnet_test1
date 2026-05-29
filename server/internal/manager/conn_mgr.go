package manager

import (
	"gnet_test1/internal/protocol"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/gnet/v2"
)

type IConnection interface {
	ID() uint64
	RemoteAddr() string
	Send(cmdID uint32, body []byte) error
	Close()
	Context() any
	SetContext(ctx any)
	PendingTasks() int32
	AddPendingTask() int32
	DelPendingTask() int32
}

type gnetConn struct {
	conn         gnet.Conn
	id           uint64
	pendingTasks int32
	ctx          any
}

func (c *gnetConn) ID() uint64 {
	return c.id
}

func (c *gnetConn) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *gnetConn) Send(cmdID uint32, body []byte) error {
	return protocol.SendPacket(c.conn, cmdID, body)
}

func (c *gnetConn) Close() {
	c.conn.Close()
}

func (c *gnetConn) Context() any {
	return c.ctx
}

func (c *gnetConn) SetContext(ctx any) {
	c.ctx = ctx
}

func (c *gnetConn) PendingTasks() int32 {
	return c.pendingTasks
}

func (c *gnetConn) AddPendingTask() int32 {
	return atomic.AddInt32(&c.pendingTasks, 1)
}

func (c *gnetConn) DelPendingTask() int32 {
	return atomic.AddInt32(&c.pendingTasks, -1)
}

type ConnManager struct {
	conns      map[uint64]IConnection
	mu         sync.RWMutex
	nextConnID uint64
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make(map[uint64]IConnection),
	}
}

func (cm *ConnManager) Add(c IConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[c.ID()] = c
}

func (cm *ConnManager) Remove(id uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conns, id)
}

func (cm *ConnManager) Get(id uint64) IConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conns[id]
}

func (cm *ConnManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.conns)
}

func (cm *ConnManager) NextID() uint64 {
	return atomic.AddUint64(&cm.nextConnID, 1)
}

func NewGnetConn(conn gnet.Conn, id uint64) IConnection {
	gc := &gnetConn{
		conn: conn,
		id:   id,
	}
	conn.SetContext(gc)
	return gc
}
