package main

import "fmt"

func main() {
	fmt.Println(LeastNum([]int{4, 5, 1, 6, 2, 7, 3, 8}, 4))
}

//给定一个长度为n的可能重复的数组，找出其中不去重复的最小k个数
//比如 4，5，1，6，2，7，3，8 这个8个数，最小的4位数1，2，3，4

func LeastNum(input []int, k int) []int {
	if k == 0 || k > len(input) {
		return []int{}
	}

	newNum := quickSort(input)
	return newNum[:k]
}

func quickSort(input []int) []int {
	if len(input) < 2 {
		return input
	}

	pivot := input[0]
	left := []int{}
	right := []int{}
	for _, v := range input[1:] {
		if v <= pivot {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	//input = append(quickSort(left), pivot)
	//input = append(input, quickSort(right)...)
	return append(append(quickSort(left), pivot), quickSort(right)...)
}
