package main

import "fmt"

// 用两个栈实现一个队列，先进先出，push插入，pop删除
var (
	stack1 []int //在第一个栈插入
	stack2 []int //在第二个栈删除
)

func main() {
	push(1)
	push(2)
	push(3)

	fmt.Println(pop())
	fmt.Println(pop())
}
func push(node int) {
	stack1 = append(stack1, node)
}

func pop() int {
	if len(stack2) == 0 {
		for len(stack1) > 0 {
			stack2 = append(stack2, stack1[len(stack1)-1])
			stack1 = stack1[:len(stack1)-1]
		}
	}

	if len(stack2) == 0 {
		return -1
	}
	res := stack2[len(stack2)-1]
	stack2 = stack2[:len(stack2)-1]
	return res
}
