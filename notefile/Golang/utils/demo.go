package utils

import (
	"log"
	"time"
)

func Run() {
	batch := NewBatchWithOptions(
		0,
		func(items []map[string]any) {
			dump(items)
		},
	)

	data := []map[string]any{
		{"id": 1, "name": "alice", "score": 95},
		{"id": 2, "name": "bob", "score": 87},
	}

	for {
		time.Sleep(time.Millisecond * 50)
		// data 可以从数据库读取，此处模拟
		for _, val := range data {
			batch.Add(val)
		}
		batch.FlushSynchronous()
	}

}

func dump(items []map[string]any) {
	return
}

func Run2() {
	lefetime := time.Minute
	interval := time.Second * 3
	ticker := time.NewTicker(interval)
	done := make(chan bool)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("ticker ticked")
			case <-done:
				log.Println("ticker stopped")
				return
			}
		}
	}()

	time.Sleep(lefetime)
	done <- true
	log.Println("done")
}
