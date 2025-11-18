package listnode

import "testing"

func TestLRU(t *testing.T) {
	lru := NewLRUCache(3)
	lru.Put(1, 100)
	lru.Put(2, 200)
	lru.Put(3, 300)

	lru.Print()
}
