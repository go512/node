package utils

import "time"

func Run()  {
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
		//data 可以从数据库读取，此处模拟
		for _, val := range data {
			batch.Add(val)
		}
		batch.FlushSynchronous()
	}

}

func dump(items []map[string]any)  {
	return
}