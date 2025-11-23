package pkg

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func Test_ActivityService(t *testing.T) {
	var count int64
	var wg sync.WaitGroup

	//启动1000个goroutine， 每个对count+1
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&count, 1) // 原子操作
		}()
	}
	wg.Wait()
	fmt.Println("count:", count)

	atomic.StoreInt64(&count, 0)
	fmt.Println("flag:", atomic.LoadInt64(&count))

	var config atomic.Value
	config.Store(map[string]string{"server": "localhost", "port": "8080"})
	//原子读取
	cfg := config.Load().(map[string]string)
	fmt.Println(cfg["server"])
	config.Store(map[string]string{"server": "192.168.1.1", "port": "9000"})
	cfg2 := config.Load().(map[string]string)
	fmt.Println(cfg2["server"])
}
