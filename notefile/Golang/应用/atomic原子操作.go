package 应用

import (
	"fmt"
	"sync"
	"sync/atomic"
)

/**
 go语言中 sync/atomic 包提供了底层原子操作， 用于实现并发安全的变量访问，无需使用互斥锁
常用操作类型：
整型： int32/ int64 / uint32 / uint64 / uintptr
指针： unsafe.Pointer
自定义类型： atomic.Value (任意类型)

2、加载/存储操作（Load/Store）
func (v *Value) Store(x any)
Loadxxx: 原子读取变量值  x为nil会panic
Storexxx: 原子写入变量值,如果未调用store返回nil

3、交换操作（Swap）原子替换变量值，返回旧值
func SwapInt64(add *int64, new int64) (old int64)

4、atomic.value （任意类型原子操作）
用于存储任意类型的值，支持原子加载/存储交换

5、使用场景 并发计数器
type Counter struct {
	count int64
}

func (c *Counter) Inc() {
	atomic.AddInt64(&c.count, 1)
}

func (c *Counter) Dec() {
	atomic.AddInt64(&c.count, -1)
}

func (c *Counter) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

*/

func test_main() {
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

	oldValue := atomic.SwapInt64(&count, 100)
	fmt.Println("oldValue: ", oldValue, "newValue:", count)

	var config atomic.Value
	config.Store(map[string]string{"server": "localhost", "port": "8080"})
	//原子读取
	cfg := config.Load().(map[string]string)
	fmt.Println(cfg["server"])
	config.Store(map[string]string{"server": "192.168.1.1", "port": "9000"})
	fmt.Println(cfg["server"])
}
