package pool

import (
	"sync"

	"github.com/valyala/bytebufferpool"
)

var ByteBufferPool bytebufferpool.Pool

func GetByteBuffer() *bytebufferpool.ByteBuffer {
	return ByteBufferPool.Get()
}

func PutByteBuffer(buf *bytebufferpool.ByteBuffer) {
	ByteBufferPool.Put(buf)
}

type BytesPool struct {
	pool sync.Pool
}

func NewBytesPool(minSize int) *BytesPool {
	return &BytesPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, minSize)
			},
		},
	}
}

func (p *BytesPool) Get(size int) []byte {
	b := p.pool.Get().([]byte)
	if cap(b) < size {
		return make([]byte, size)
	}
	return b[:size]
}

func (p *BytesPool) Put(b []byte) {
	p.pool.Put(b)
}

var defaultPool = NewBytesPool(1024)

func GetBytes(size int) []byte {
	return defaultPool.Get(size)
}

func PutBytes(b []byte) {
	defaultPool.Put(b)
}

var WorkTaskPool = sync.Pool{
	New: func() interface{} {
		return &WorkTask{}
	},
}

func GetWorkTask() *WorkTask {
	return WorkTaskPool.Get().(*WorkTask)
}

func PutWorkTask(task *WorkTask) {
	if task.Body != nil {
		PutBytes(task.Body)
	}
	task.ConnID = 0
	task.CmdID = 0
	task.Body = nil
	task.Conn = nil
	WorkTaskPool.Put(task)
}
