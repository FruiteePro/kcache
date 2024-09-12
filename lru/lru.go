package lru

import (
	"container/list"
	"log"
	"math/rand"
	"time"
)

/*
lru cache 结构体，用来实现 lru 缓存淘汰算法。
maxBytes：最大存储容量
nBytes：已占用的容量
ll：直接使用 Go 语言标准库实现的双向链表list.List，双向链表常用于维护缓存中各个数据的访问顺序，以便在淘汰数据时能够方便地找到最近最少使用的数据。
cache：map,键是字符串，值是双向链表中对应节点的指针
OnEvicted：是某条记录被移除时的回调函数，可以为 nil
defaultTTL：记录在缓存中的默认过期时间
*/

type LRUCache struct {
	maxBytes   int64
	nBytes     int64
	ll         *list.List
	cache      map[string]*list.Element
	OnEvicted  func(key string, value Value)
	defaultTTL time.Duration
}

type entry struct {
	key    string
	value  Value
	expire time.Time
}

type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(key string, value Value), defaultTTL time.Duration) *LRUCache {
	return &LRUCache{
		maxBytes:   maxBytes,
		nBytes:     0,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
		OnEvicted:  OnEvicted,
		defaultTTL: defaultTTL,
	}
}

func (c *LRUCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		if kv.expire.Before(time.Now()) {
			c.RemoveElement(ele)
			log.Printf("The LRUCache key-%s has expired", key)
			return nil, false
		}
		c.ll.MoveToFront(ele)
		return kv.value, true
	}
	return nil, false
}

func (c *LRUCache) RemoveElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *LRUCache) RemoveOldest() {
	for e := c.ll.Back(); e != nil; e = e.Prev() {
		kv := e.Value.(*entry)
		if kv.expire.Before(time.Now()) {
			c.RemoveElement(e)
			break
		}
	}
	e := c.ll.Back()
	c.RemoveElement(e)
}

func (c *LRUCache) Add(key string, value Value, ttl time.Duration) {
	expireTime := time.Now().Add(ttl + time.Duration(rand.Intn(60))*time.Second)
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		if kv.expire.Before(time.Now()) {
			kv.expire = expireTime
		}
	} else {
		ele = c.ll.PushFront(&entry{key, value, expireTime})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}
