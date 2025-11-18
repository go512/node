package listnode

import "fmt"

/**
 *  双链表 实现的lru
 *  ｜------｜------｜-------｜------>  ｜-----｜----｜------｜
 *  ｜ prev ｜ data ｜ next  ｜	 	   ｜ pre ｜data｜ next ｜
 *  ｜------｜------｜-------｜ <-----   ｜----｜-----｜-----｜
 *  使用哈希表+双向链表实现lru
 *  定义双向链表节点，包含key value，前驱节点和后继节点
 *  初始化LRU缓存 ，设置容量
 */

// 双向链表节点
type node struct {
	key   int
	value int
	prev  *node //前驱节点
	next  *node //后继节点
}

// LRU缓存
type LRUCache struct {
	cap   int
	cache map[int]*node
	head  *node
	tail  *node
}

func NewLRUCache(cap int) *LRUCache {
	//初始化头尾哨兵节点 简化边界处理
	head := &node{}
	tail := &node{}
	head.next = tail
	tail.prev = head

	return &LRUCache{
		cap:   cap,
		cache: make(map[int]*node, cap),
		head:  head,
		tail:  tail,
	}
}

// 辅助函数， 移除制定节点
func (this *LRUCache) removeNode(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

//辅助函数， 添加节点到头部
/**
 * 1、保持链表完整性: 首先要把新的节点插入到头节点的下一个节点
 * 2、然后在把头节点和新插入的节点调换位置
 */
func (this *LRUCache) addToHead(n *node) {
	//先插入头节点的下一个节点位置
	n.prev = this.head
	n.next = this.head.next
	//再把头节点和插入的节点调换位置
	this.head.next.prev = n
	this.head.next = n
}

// 移动节点到头部
func (this *LRUCache) moveToHead(n *node) {
	this.removeNode(n)
	this.addToHead(n)
}

// 移除尾节点（返回被删除的节点）
/**
head <-> node1 <-> node2 <-> ... <-> last_node <-> tail
                                      ↑              ↑
                                   tail.prev      尾哨兵
lru.tail 是尾部哨兵节点（dummy node），不是实际存储数据的节点
lru.tail.prev 指向的是最后一个有效节点，也就是真正的尾节点
*/
func (this *LRUCache) removeTail() *node {
	tail := this.tail.prev
	this.removeNode(tail)
	return tail
}

// 获取值
func (this *LRUCache) Get(key int) int {
	if n, exists := this.cache[key]; exists {
		//如果存在，则把该节点移到头部
		this.moveToHead(n)
		return n.value
	}
	return -1
}

func (this *LRUCache) Put(key int, value int) {
	if n, exists := this.cache[key]; exists {
		//如果存在，则把该节点移到头部
		n.value = value
		this.moveToHead(n)
		return
	}
	if len(this.cache) >= this.cap {
		tail := this.removeTail()
		delete(this.cache, tail.key)
	}

	newNode := &node{
		key:   key,
		value: value,
	}
	this.cache[key] = newNode
	this.addToHead(newNode)
}

func (this *LRUCache) Size() int {
	return len(this.cache)
}

func (this *LRUCache) Print() {
	fmt.Printf("ManualLRU Cache (capacity: %d, size: %d): ", this.cap, this.Size())
	for curr := this.head.next; curr != this.tail; curr = curr.next {
		fmt.Printf("[%d:%d] ", curr.key, curr.value)
	}
	fmt.Println()
}
