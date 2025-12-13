package chann

import "sync"

//SliceSplit[S ~[]E, E any] 是一个泛型函数，用于将切片 S 分割为多个子切片，每个子切片的长度为 n
// 如 []int/ []string/ []struct{} 解决批量处理大切片的场景 （如批量插入数据库，分批次发送请求）

/**
A: 边界处理
	1、边界处理 切片为空/分块数 <= 1 时，直接返回原切片(不分割)
	2、切片长度 <= 分块数 -> 每个元素单独成一个子切片
B: 均匀分块
	1、计算基础快大小 base = 总长度 / 分块数
	2、计算余数 extra = 总长度 % 分块数
	2、计算最后一个块的长度 last = 总长度 - base * (分块数 - 1)
C: 泛型支持
	1、通过 S ~[]E, E any 适配所有切片类型，无类型限制。
*/

func SliceSplit[S ~[]E, E any](s S, parts int) [][]E {
	//边界1 分块数 <= 1 时，直接返回原切片(不分割)
	n := len(s)
	if parts <= 1 || n <= 0 {
		return [][]E{s}
	}

	//边界2 切片长度 <= 分块数 -> 每个元素单独成一个子切片
	if n <= parts {
		result := make([][]E, n)
		for i := range s {
			result[i] = s[i : i+1] // 注意这里的 i+1 是切片的右边界，不是长度
			//fmt.Println(i, result[i])
		}
		return result
	}

	base := n / parts
	extra := n % parts
	left := 0
	result := make([][]E, parts)

	for i := 0; i < parts; i++ {
		size := base
		if i < extra {
			size++
		}

		right := left + size
		result[i] = s[left:right]
		left = right // 更新左边界
		//fmt.Println(i, result[i])
	}
	return result
}

// SliceSplitBySize 根据大小分块
// SliceSplitBySize(intSlice, 3) → [[1 2 3] [4 5 6] [7 8 9] [10]]
func SliceSplitBySize[S ~[]E, E any](s S, size int) [][]E {
	if len(s) == 0 || size <= 0 {
		return [][]E{s}
	}

	var result [][]E
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}

		result = append(result, s[i:end])
	}
	return result
}

// ProcessInParallel 并发处理分块切片
func ProcessInParallel[S ~[]E, E any](s S, parrts int, fn func(sub []E)) {
	chunks := SliceSplit(s, parrts)
	wg := sync.WaitGroup{}
	for _, chunk := range chunks {
		wg.Add(1)
		go func(sub []E) {
			defer wg.Done()
			fn(sub)
		}(chunk)
	}
	wg.Wait()
}
