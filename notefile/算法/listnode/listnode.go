package listnode

type ListNode struct {
	Val  int
	Next *ListNode
}

// 1、删除有序列表中重复的元素
// 给出的链表为  1→1→2,返回 1→2
func deleteDup(head *ListNode) *ListNode {
	if head == nil {
		return head
	}

	//双指针
	slow, fast := head, head.Next
	for fast != nil {
		//快慢指针相同，则慢指针指向快指针
		if slow.Val == fast.Val {
			slow.Next = fast.Next
		} else {
			//否则向后滑动
			slow = slow.Next
		}
		fast = fast.Next
	}

	return head
}
