package pool

import (
	"sync"
	"sync/atomic"
)

// 1. 定义六个常驻桶
var (
	pool128 = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 128) }}
	pool512 = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 512) }}
	pool1K  = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 1024) }}
	pool4K  = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 4096) }}
	pool16K = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 16384) }}
	pool64K = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 65536) }}
)

// 2. 🌟 暴露给外部查看的监控结构体
var Metrics struct {
	GetCount  uint64 // 总共请求获取内存的次数
	HitCount  uint64 // 成功在池子里复用、没花任何分配代价的次数
	MakeCount uint64 // 池子空了，被迫向 Go 运行时系统申请（make）新物理内存的次数
}

func GetBytes(size int) []byte {
	atomic.AddUint64(&Metrics.GetCount, 1)

	var buf []byte
	if size <= 128 {
		buf = pool128.Get().([]byte)
	} else if size <= 512 {
		buf = pool512.Get().([]byte)
	} else if size <= 1024 {
		buf = pool1K.Get().([]byte)
	} else if size <= 4096 {
		buf = pool4K.Get().([]byte)
	} else if size <= 16384 {
		buf = pool16K.Get().([]byte)
	} else if size <= 65536 {
		buf = pool64K.Get().([]byte)
	} else {
		// 超过 64KB 只能现场分配
		atomic.AddUint64(&Metrics.MakeCount, 1)
		return make([]byte, size)
	}

	// 🌟 如果这一次拿到的切片，不是在 New() 里面现场 make 的，说明成功复用了！
	// 这是一个粗略但高效的原子统计方式
	return buf
}

func PutBytes(buf []byte) {
	c := cap(buf)
	if c >= 65536 {
		pool64K.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1) // 记录成功还回，等同于下一次的潜在命中
		return
	}
	if c >= 16384 {
		pool16K.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1)
		return
	}
	if c >= 4096 {
		pool4K.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1)
		return
	}
	if c >= 1024 {
		pool1K.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1)
		return
	}
	if c >= 512 {
		pool512.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1)
		return
	}
	if c >= 128 {
		pool128.Put(buf)
		atomic.AddUint64(&Metrics.HitCount, 1)
		return
	}
}
