package utils

import (
	"math"
	"sync"
	"time"
)

const (
	defaultBatchSize = 1000
)

type Batch[T any] struct {
	l             sync.RWMutex  // 读写锁， 保证Add 和 Flush并发安全
	flushInterval time.Duration // 后台自动刷新的间隔 （0 = 不自动刷）
	size          int           // 当前缓冲区已有元素个数
	maxSize       int           // 缓冲区最大容量 （默认 defaultBatchSize）
	maxCommitSize int           // 每次写库的分快上限

	list                      []T             // 预分配的缓冲数组
	flushOp                   func(items []T) // 落库回调
	runningBackgroundFlushOps int             // 进行中的异步刷盘数
}

// 设置maxSize
func WithMaxSize[T any](maxSize int) func(b *Batch[T]) {
	return func(b *Batch[T]) {
		b.maxSize = maxSize
		if b.maxCommitSize == 0 || b.maxCommitSize > maxSize {
			b.maxCommitSize = maxSize
		}
	}
}

func WithMaxCommitSize[T any](maxCommitSize int) func(b *Batch[T]) {
	return func(b *Batch[T]) {
		b.maxCommitSize = maxCommitSize
	}
}

// Batch[T] 可以看作是一个struct，比如*Batch[Feed] feed类型的结构体
// 返回*Batch[T]指针式因为多个协程都在操作，主循环goroutine调用batch.Add,后台协程操作保存
func NewBatchWithOptions[T any](flushInterval time.Duration, flushOp func(items []T), options ...func(b *Batch[T])) *Batch[T] {
	maxSize := defaultBatchSize
	maxCommitSize := maxSize

	b := &Batch[T]{
		flushInterval: flushInterval,
		flushOp:       flushOp,
		maxCommitSize: maxCommitSize,
	}

	for _, option := range options {
		option(b)
	}

	if b.maxSize == 0 {
		b.maxSize = maxSize
	}

	b.list = make([]T, b.maxSize)

	// 大于0启动异步
	if b.flushInterval > 0 {
		go func() {
			for {
				time.Sleep(b.flushInterval)
				b.flushAsynchronous()
			}
		}()
	}

	return b
}

// 异步保存
func (b *Batch[T]) flushAsynchronous() {
	b.l.Lock()

	n := b.size
	if n == 0 {
		b.l.Unlock()
		return
	}

	commitSize := n
	if commitSize > b.maxCommitSize {
		commitSize = b.maxCommitSize
	}

	items := make([]T, commitSize)
	copy(items, b.list[:commitSize])

	// 将剩余元素前移，并清空尾部避免内存泄漏
	remaining := n - commitSize
	if remaining > 0 {
		copy(b.list, b.list[commitSize:n])
		clear(b.list[remaining:n])
	}
	b.size = remaining

	b.runningBackgroundFlushOps++
	b.l.Unlock()

	go func() {
		defer func() {
			b.l.Lock()
			b.runningBackgroundFlushOps--
			b.l.Unlock()
		}()
		b.flushOp(items)
	}()
}

// 追加到list缓冲
func (b *Batch[T]) Add(items ...T) {
	b.l.Lock()
	defer b.l.Unlock()

	for i, item := range items {
		b.list[b.size+i] = item
	}

	b.size += len(items)
}

func min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

// 清空缓存
func (b *Batch[T]) resetBuffer() {
	b.list = make([]T, b.maxSize)
	b.size = 0
}

// 分批把缓存的数据入库
func (b *Batch[T]) FlushSynchronous() {
	for i := 0; i < b.size; i += b.maxCommitSize {
		b.flushOp(b.list[i:min(i+b.maxCommitSize, b.size)])
	}
	b.resetBuffer()
}
