package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		//模拟一些工作
		time.Sleep(20 * time.Second)
		cancel()
	}()

	//等待取消信号
	for { //死循环
		select {
		case <-ctx.Done():
			fmt.Println("上下文被取消了", ctx.Err())
		case <-time.After(5 * time.Second): // 每5秒会检测一次
			fmt.Println("超时了")
		}
	}
}
