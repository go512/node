package main

import "fmt"

func main() {
	push(1)
	push(2)
	push(3)

	fmt.Println(pop())
	fmt.Println(pop())
	fmt.Println(sortA())
	fmt.Println(isValid("[]{}"))
}

// 用两个栈实现一个队列，先进先出，push插入，pop删除
var (
	stack1 []int //在第一个栈插入
	stack2 []int //在第二个栈删除
)

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

// 栈A元素是乱序的，利用一个栈B和最多三个变量对A进行排序
func sortA() []int {
	//模拟乱序栈A
	stackA := []int{5, 2, 3, 1, 4}
	var stackB []int
	var temp int

	//把A元素弹出，并判断是否比栈B的栈顶元素小，如果是则入栈B，否则入栈A
	for len(stackA) > 0 {
		temp = stackA[len(stackA)-1]
		stackA = stackA[:len(stackA)-1]

		for len(stackB) > 0 && temp > stackB[len(stackB)-1] {
			stackA = append(stackA, stackB[len(stackB)-1])
			stackB = stackB[:len(stackB)-1]
		}

		//把temp压入栈B
		stackB = append(stackB, temp)
	}
	//B降序，倒入A升序
	for len(stackB) > 0 {
		stackA = append(stackA, stackB[len(stackB)-1])
		stackB = stackB[:len(stackB)-1]
	}

	return stackA
}

// 3、有效的括号，用右边的映射，左边的入栈
// ()[]{}
func isValid(s string) bool {
	//创建映射关系，右括号对应左括号
	pairs := map[byte]byte{
		')': '(',
		']': '[',
		'}': '{',
	}

	//使用切片模拟栈
	stack := []byte{}
	for _, char := range s {
		if left, ok := pairs[byte(char)]; ok {
			if len(stack) == 0 || stack[len(stack)-1] != left {
				return false
			}
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, byte(char))
		}
	}

	return len(stack) == 0
}
