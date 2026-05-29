package pool

import (
	"sync"
	"sync/atomic"
)

// 定义六个常驻桶
var (
	// 引入一个内部标记，用来判断是否触发了 New
	pool128 = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 128) }}
	pool512 = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 512) }}
	pool1K  = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 1024) }}
	pool4K  = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 4096) }}
	pool16K = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 16384) }}
	pool64K = sync.Pool{New: func() any { atomic.AddUint64(&Metrics.MakeCount, 1); return make([]byte, 65536) }}
)

// Metrics 暴露给外部查看的监控结构体
var Metrics struct {
	GetCount  uint64 // 总共请求获取内存的次数
	HitCount  uint64 // 真正从池子中复用成功的次数
	MakeCount uint64 // 池子空了，被迫向系统申请新物理内存的次数
	RawAlloc  uint64 // 超过 64KB 导致不得不现场分配的次数
}

func GetBytes(size int) []byte {
	atomic.AddUint64(&Metrics.GetCount, 1)

	var buf []byte
	// 记录进入分配前的 MakeCount 数量
	beforeMake := atomic.LoadUint64(&Metrics.MakeCount)

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
		// 超过 64KB 只能现场分配，不进池子
		atomic.AddUint64(&Metrics.RawAlloc, 1)
		return make([]byte, size)
	}

	// 🌟 精准判断是否命中：如果 MakeCount 没有增加，说明是从池子里捞出来的
	//  这里可能会统计错误，但比例较少，不影响监控指标
	afterMake := atomic.LoadUint64(&Metrics.MakeCount)
	if afterMake == beforeMake {
		atomic.AddUint64(&Metrics.HitCount, 1)
	}

	// 🌟 核心修复：切片重置。将长度缩至用户需要的 size，保留其原本的 capacity
	return buf[:size]
}

func PutBytes(buf []byte) {
	// 释放切片引用，防止底层数组还回去后，里面的旧数据导致内存泄漏（尤其是如果里面存了指针）
	// 这里选择直接还回，但限制只有原始容量匹配的才放回，防止池子内的对象扩容变形
	c := cap(buf)

	if c == 65536 {
		pool64K.Put(buf[:c]) // 还原长度为完整容量再放回
		return
	}
	if c == 16384 {
		pool16K.Put(buf[:c])
		return
	}
	if c == 4096 {
		pool4K.Put(buf[:c])
		return
	}
	if c == 1024 {
		pool1K.Put(buf[:c])
		return
	}
	if c == 512 {
		pool512.Put(buf[:c])
		return
	}
	if c == 128 {
		pool128.Put(buf[:c])
		return
	}

	// 如果容量不对（比如被外部 append 扩容过了），直接丢弃让 GC 回收，确保池子纯净
}
