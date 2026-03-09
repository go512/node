package chann

import (
	"fmt"
	"sync"
	"time"
)

type Job func()

type Pool struct {
	jobs    chan Job
	workers int
	wg      sync.WaitGroup
}

func NewPool(workers int, jobQueueSize int) *Pool {
	pool := &Pool{
		jobs:    make(chan Job, jobQueueSize),
		workers: workers,
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			for job := range pool.jobs {
				job()
			}
		}()
	}
	return pool
}

func (p *Pool) Add(job Job) {
	p.jobs <- job
	if len(p.jobs) == 10 {
		fmt.Println("---------------")
		time.Sleep(1 * time.Second)
	}
}

func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
}

//-----------------------

func processFile(filename string) func() {
	return func() {
		fmt.Println("处理文件：", filename)
		//模拟耗时操作
	}
}
