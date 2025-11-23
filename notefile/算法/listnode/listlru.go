package listnode

import (
	"container/list"
	"fmt"
	"sync"
)

//LRU（最近最少使用）缓存示例
//基础版LRU缓存（非线程安全）

type LRUCacheV2 struct {
	cap   int
	cache map[any]*list.Element
	list  *list.List
}

// entry 链表节点存储的键值对
type entry struct {
	key   any
	value interface{}
}

func NewLRUCacheV2(cap int) (*LRUCacheV2, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("cap must be greater than 0")
	}

	return &LRUCacheV2{
		cap:   cap,
		cache: make(map[any]*list.Element),
		list:  list.New(),
	}, nil
}

// Get  获取缓存值 不存在返回nil
func (this *LRUCacheV2) Get(key any) interface{} {
	if elem, ok := this.cache[key]; ok {
		//将访问的节点移动到链表头部（标记为最近使用）
		this.list.MoveToFront(elem)
		return elem.Value.(*entry).value
	}
	return nil
}

// put 添加/更新缓存（容量满时淘汰醉酒未使用的条目）
func (this *LRUCacheV2) Put(key, value any) {
	//若键已存在，更新值并移动到头部
	if elem, ok := this.cache[key]; ok {
		elem.Value.(*entry).value = value
		this.list.MoveToFront(elem)
		return
	}

	//容量已满,删除最久未使用的条目（链表尾部）
	if this.list.Len() >= this.cap {
		backElem := this.list.Back()
		if backElem != nil {
			delete(this.cache, backElem.Value.(*entry).key)
			this.list.Remove(backElem)
		}
	}

	//添加新的条目到链表头部
	newElem := this.list.PushFront(&entry{key, value})
	this.cache[key] = newElem
}

// Remove 删除置顶建的缓存
func (this *LRUCacheV2) Remove(key any) {
	if elem, ok := this.cache[key]; ok {
		this.list.Remove(elem)
		delete(this.cache, key)
	}
}

func (this *LRUCacheV2) Len() int {
	return this.list.Len()
}

func (this *LRUCacheV2) Clear() {
	this.list.Init()
	this.cache = make(map[any]*list.Element)
}

//线程安全版本，支持并发

type SyncLRUCache struct {
	lru   *LRUCacheV2
	mutex sync.RWMutex
}

func NewSyncLRUCache(cap int) (*SyncLRUCache, error) {
	lru, err := NewLRUCacheV2(cap)
	if err != nil {
		return nil, err
	}

	return &SyncLRUCache{
		lru:   lru,
		mutex: sync.RWMutex{},
	}, nil
}

// Get 线程安全的获取操作
func (s *SyncLRUCache) Get(key any) any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lru.Get(key)
}

// put 线程安全的添加/更新
func (s *SyncLRUCache) Put(key, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lru.Put(key, value)
}

func (s *SyncLRUCache) Remove(key any) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.lru.Remove(key)
}

func (s *SyncLRUCache) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lru.Len()
}

func (s *SyncLRUCache) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lru.Clear()
}
