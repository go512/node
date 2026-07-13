package Tc

import (
	"fmt"
	"notefile-golang/utils"
)

func LoadAllDoc[T TsDoc](sportId int32, subject int32) {
	batch := utils.NewBatchWithOptions(
		0,
		func(items []T) {
			// dump(items)
		},
		utils.WithMaxSize[T](1000),
		utils.WithMaxCommitSize[T](1000),
	)

	url := getEndpointUrl(subject, sportId)
	// todo
	// 这里请求接口获取数据
	fmt.Println(url)

	// batch.Add([]T)
	batch.FlushSynchronous()
}

func getEndpointUrl(subject, sportId int32) string {
	var url string
	switch subject {
	case 1:
		url = fmt.Sprintf("https://api.sportradar.us/%d/%d/en/matches.json?api_key=%s", subject, sportId, 1)
	case 2:
		url = fmt.Sprintf("https://api.sportradar.us/%d/%d/en/matches.json?api_key=%s", subject, sportId, 1)
	default:
		url = fmt.Sprintf("https://api.sportradar.us/%d/%d/en/matches.json?api_key=%s", subject, sportId, 1)
	}
	return url
}
