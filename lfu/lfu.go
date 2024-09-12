package lfu

import (
	"container/heap"
	"log"
	"time"
)

/*
LFUCache 定义了一个结构体，用来实现lfu缓存淘汰算法
maxBytes：最大存储容量
nBytes：已占用的容量
heap：使用一个 heap 来管理缓存项，heap 中的元素按照频率排序(heap实现了一个最小堆，即堆顶元素是最小值)
cache：map，键是字符串，值是堆中对应节点的指针
OnEvicted：是某条记录被移除时的回调函数，可以为 nil
defaultTTL：记录在缓存中的默认过期时间
*/

type LFUCache struct {
	maxBytes   int64
	nBytes     int64
	heap       *entryHeap
	cache      map[string]*entry
	OnEvicted  func(key string, value Value)
	defaultTTL time.Duration
}

type Value interface {
	Len() int
}

type entry struct {
	key    string
	value  Value
	freq   int
	index  int
	expire time.Time
}

// entryHeap 实现 heap.Interface 接口

type entryHeap []*entry

func (h entryHeap) Len() int {
	return len(h)
}

func (h entryHeap) Less(i, j int) bool {
	return h[i].freq < h[j].freq
}

func (h entryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *entryHeap) Push(x interface{}) {
	entry := x.(*entry)
	entry.index = len(*h)
	*h = append(*h, entry)
}

func (h *entryHeap) Pop() interface{} {
	old := *h
	n := len(*h)
	entry := *old[n-1]
	entry.index = -1
	*h = old[0 : n-1]
	return entry
}

func New(maxBytes int64, onEvicted func(key string, value Value), defaultTTL time.Duration) *LFUCache {
	return &LFUCache{
		maxBytes:   maxBytes,
		heap:       &entryHeap{},
		cache:      make(map[string]*entry),
		OnEvicted:  onEvicted,
		defaultTTL: defaultTTL,
	}
}

func (c *LFUCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		if ele.expire.Before(time.Now()) {
			c.removeElement(ele)
			log.Printf("The LFUcache key-%s has expired", key)
			return nil, false
		}
		ele.freq++
		//Fix 方法用于在索引 index 处的元素值发生变化后重新确立堆的顺序。在索引 index 的元素值发生改变后，调用 Fix 方法可以保持堆的性质。
		//Fix 方法的时间复杂度是 O(log n)，其中 n = h.Len() 表示堆中元素的数量。
		heap.Fix(c.heap, ele.index)
		return ele.value, true
	} else {
		return nil, false
	}
}

func (c *LFUCache) RemoveOldest() {
	entry := heap.Pop(c.heap).(*entry)
	delete(c.cache, entry.key)
	c.nBytes -= int64(len(entry.key)) + int64(entry.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(entry.key, entry.value)
	}
}

func (c *LFUCache) Add(key string, value Value, ttl time.Duration) {
	if ele, ok := c.cache[key]; ok {
		ele.value = value
		ele.freq++
		ele.expire = time.Now().Add(ttl)
		heap.Fix(c.heap, ele.index)
	} else {
		entry := &entry{
			key:    key,
			value:  value,
			freq:   1,
			expire: time.Now().Add(ttl),
		}
		c.cache[key] = entry
		heap.Push(c.heap, entry)
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *LFUCache) Len() int {
	return len(c.cache)
}

func (c *LFUCache) removeElement(e *entry) {
	heap.Remove(c.heap, e.index)
	delete(c.cache, e.key)
	c.nBytes -= int64(len(e.key)) + int64(e.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(e.key, e.value)
	}
}
