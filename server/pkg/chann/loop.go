package chann

import (
	"fmt"
	"sync"
)

type MatchStatusChange struct {
	MatchId          string            `json:"match_id"`
	SportEventStatus *SportEventStatus `json:"sport_event_status"`
}

type SportEventStatus struct {
	Id        string `json:"id"`
	Status    string `json:"status"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
}

func HandleMatchStatusChange() {
	cuts := [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	// 创建一个带缓冲的通道，容量为cuts的总和
	dataCh := make(chan *MatchStatusChange, 7)
	// 创建一个doneCh通道，用于通知所有goroutine完成
	doneCh := make(chan struct{})
	go func() {
		for data := range dataCh {
			fmt.Println("---- ", data.MatchId)
		}
		// 所有goroutine完成后关闭doneCh通道
		close(doneCh)
	}()

	var wg sync.WaitGroup
	for _, cut := range cuts {
		wg.Go(func() {
			for _, c := range cut {
				dataCh <- &MatchStatusChange{
					MatchId: fmt.Sprintf("match_%d", c),
					SportEventStatus: &SportEventStatus{
						Id:     fmt.Sprintf("sport_event_%d", c),
						Status: "status",
					},
				}
			}
		})
	}

	wg.Wait()
	close(dataCh)
	<-doneCh
}
