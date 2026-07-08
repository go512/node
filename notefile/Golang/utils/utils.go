package utils

import (
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
	maxCommitSize int

	list                      []T
	flushOp                   func(items []T)
	runningBackgroundFlushOps int
}

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

func NewBatch() {

}
