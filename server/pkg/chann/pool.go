package chann

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task 定义任务接口：包含超时执行+ 成功/失败回调
type Task interface {
	Execute(ctx context.Context) (any, error) //支持超时控制
	OnComplete(result any)                    //成功回调
	OnError(err error)                        //失败回调
}

// workerPool 协程池结构体
type WorkerPool struct {
	taskQueue chan Task //任务队列
	poolSize  int       //最大并发数
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc

	//监控指标 （原子操作保证并发安全）
	totalTasks    uint64 //总任务数
	successTasks  uint64 //成功数
	failureTasks  uint64
	totalExecTime int64 //总执行时间
}

/*
 *	创建协程池
 *  poolSize: 最大并发数
 * taskQueueSize: 任务队列缓冲大小
 * taskTimeout: 单个任务默认超时时间，（可在任务层覆盖）
 */
func NewWorkerPool(poolSize, taskQueueSize int, taskTimeout time.Duration) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		taskQueue: make(chan Task, taskQueueSize),
		poolSize:  poolSize,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// GetStats 获取协程池监控指标
func (wp *WorkerPool) GetStats() map[string]any {
	//计算平均执行时间
	avgExecTime := 0.0
	total := atomic.LoadUint64(&wp.totalTasks)
	if total > 0 {
		totalTime := atomic.LoadInt64(&wp.totalExecTime)
		avgExecTime = float64(totalTime) / float64(total)
	}

	return map[string]any{
		"total_tasks":      total,
		"success_tasks":    atomic.LoadUint64(&wp.successTasks),
		"failure_tasks":    atomic.LoadUint64(&wp.failureTasks),
		"pool_size":        wp.poolSize,
		"queue_size":       len(wp.taskQueue),
		"queue_capacity":   cap(wp.taskQueue),
		"avg_exec_time_ms": fmt.Sprintf("%.2f", avgExecTime),
	}
}

// ResetStats 重置监控指标
func (wp *WorkerPool) ResetStats() {
	atomic.StoreUint64(&wp.totalTasks, 0)
	atomic.StoreUint64(&wp.successTasks, 0)
	atomic.StoreUint64(&wp.failureTasks, 0)
	atomic.StoreInt64(&wp.totalExecTime, 0)
}

// 向队列添加任务
func (wp *WorkerPool) AddTask(task Task) bool {
	select {
	case <-wp.ctx.Done():
		fmt.Println("协程池已关闭, 拒绝添加任务")
		return false
	case wp.taskQueue <- task:
		atomic.AddUint64(&wp.successTasks, 1)
		return true
	default:
		fmt.Println("任务队列已满， 添加失败")
		return false
	}
}

// Start 启动协程池
func (wp *WorkerPool) Start(taskTimeout time.Duration) {
	fmt.Printf("协程池启动 | 最大并发：%d | 任务队列容量：%d | 任务默认超时：%v\n",
		wp.poolSize, cap(wp.taskQueue), taskTimeout)

	for i := 0; i < wp.poolSize; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			wp.workerLoop(workerID, taskTimeout)
		}(i + 1)
	}
}

func (wp *WorkerPool) workerLoop(workerID int, defaultTimeout time.Duration) {
	fmt.Println(fmt.Sprintf("工作协程 %d 启动\n", workerID))
	for {
		select {
		case <-wp.ctx.Done():
			fmt.Printf("工作协程 %d 收到关闭信号，退出\n", workerID)
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				fmt.Printf("工作协程 %d 任务队列已关闭， 退出\n", workerID)
				return
			}

			//创建带有超时的上下文，覆盖整个任务流程
			taskCtx, taskCancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer taskCancel() //确保超时后释放资源

			//记录任务开始时间（统计执行时长）
			startTime := time.Now()
			var (
				result any
				err    error
			)

			//----------------start
			//执行任务带超时控制
			done := make(chan struct{}, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						err = errors.New(fmt.Sprintf("任务执行panic：%v", r))
					}
					done <- struct{}{}
				}()
				result, err = task.Execute(taskCtx)
			}()
			//-----------------end

			//等待任务执行完成或超时
			select {
			case <-done:
			//任务正常执行完成 （无论成功/失败）
			case <-taskCtx.Done():
				//任务超时
				err = errors.New(fmt.Sprintf("任务执行超时（超时时间：%v）", defaultTimeout))
			}

			//-------------监控指标统计
			execDuration := time.Since(startTime)
			atomic.AddInt64(&wp.totalExecTime, execDuration.Milliseconds())
			//-------------监控指标统计结束

			if err != nil {
				//任务失败
				atomic.AddUint64(&wp.failureTasks, 1)
				task.OnError(err)
			} else {
				atomic.AddUint64(&wp.successTasks, 1)
				task.OnComplete(result)
			}

			// 打印单次任务统计（可选，调试用）
			fmt.Printf("【协程%d】任务执行完成 | 耗时：%.2fms | 错误：%v\n", workerID, execDuration.Seconds(), err)
		}
	}
}

// stop 优雅关闭协程池
func (wp *WorkerPool) Stop() {
	fmt.Println("\n 开始关闭协程池")
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()

	// 打印最终监控指标
	fmt.Println("=== 协程池最终监控指标 ===")
	for k, v := range wp.GetStats() {
		fmt.Printf("%-15s: %v\n", k, v)
	}
	fmt.Println("==========================")
}
