package chann

import (
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	// 创建协程池：3个并发， 队列容量10， 单个任务默认超时1秒
	pool := NewWorkerPool(3, 10, time.Second)

	//启动协程池
	pool.Start(1 * time.Second)
	//3\定时打印监控指标
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("\n=== 实时监控指标 ===")
				stats := pool.GetStats()
				for k, v := range stats {
					fmt.Printf("%-15s: %v\n", k, v)
				}
			case <-pool.ctx.Done():
				return
			}
		}
	}()

	// 4. 添加10个测试任务（模拟不同耗时/失败场景）
	taskExecTimes := []time.Duration{
		200 * time.Millisecond,  // 成功
		300 * time.Millisecond,  // 失败（ID偶数）
		800 * time.Millisecond,  // 成功
		1200 * time.Millisecond, // 超时
		500 * time.Millisecond,  // 成功
		1500 * time.Millisecond, // 超时
		400 * time.Millisecond,  // 成功
		600 * time.Millisecond,  // 失败（ID偶数）
		700 * time.Millisecond,  // 成功
		900 * time.Millisecond,  // 失败（ID偶数）
	}

	for i := 0; i < 10; i++ {
		task := &DemoTask{
			TaskID:   i + 1,
			Content:  fmt.Sprintf("测试任务-%d", i+1),
			ExecTime: taskExecTimes[i],
		}

		// 自定义回调（任务5）
		if task.TaskID == 5 {
			task.SetOnComplete(func(result interface{}) {
				fmt.Printf("【自定义成功回调】任务5：%v → 业务处理：结果入库\n", result)
			})
		}

		// 添加任务
		if pool.AddTask(task) {
			fmt.Printf("添加任务 %d 成功（预期耗时：%v）\n", task.TaskID, task.ExecTime)
		} else {
			fmt.Printf("添加任务 %d 失败\n", task.TaskID)
		}
	}

	time.Sleep(5 * time.Second)
	pool.Stop()
}
