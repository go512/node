package yingyong

import (
	"context"
	"fmt"
	"time"
)

func selectTest(ctx context.Context) {
	//在不满足case的条件下，会一直遍历，然后走到default
	// 这样会导致空转换，导致CPU 100%
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			fmt.Println("timeout")
		default:
			//这个循环会一直空转，直到超时或者取消信号，才会退出
		}
	}

	//优化方案
	// 此时循环会退出，每5秒检测一次，但不会退出，会一直select case 不会导致cpu空转，
	// 会5秒执行一次遍历
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			fmt.Println("timeout")
		}
	}
}
