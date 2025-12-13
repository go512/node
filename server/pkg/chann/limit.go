package chann

import (
	"fmt"
	"time"
)

func LoopLimit() {
	limit := make(chan struct{}, 1)
	go func() {
		tick := time.Tick(time.Second)
		for range tick {
			select {
			case limit <- struct{}{}:
				//default:
			}
		}
	}()

	go func() {
		for {
			<-limit
			Handle("desc", 1)
		}
	}()
}

func CronTasks(second int, fn func()) {
	limit := make(chan struct{}, 1)
	go func() {
		ticker := time.Tick(time.Duration(second) * time.Second)
		for range ticker {
			select {
			case limit <- struct{}{}:
			}
		}
	}()

	go func() {
		for {
			<-limit
			fn()
		}
	}()
}

func Handle(des string, num int) {
	fmt.Println(des, time.Now().Format(time.DateTime))
	time.Sleep(time.Duration(num) * time.Second)

	return
}
