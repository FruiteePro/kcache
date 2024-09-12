package kcache

import (
	"kcache/lfu"
	"kcache/lru"
	"sync"
	"time"
)

type BaseCache interface {
	add(key string, value ByteView)
	get(key string) (value ByteView, ok bool)
}

type LRUCache struct {
	mu         sync.RWMutex  // lru的maxBytes
	lru        *lru.LRUCache // lru 结构
	cacheBytes int64         // lru的maxBytes
	ttl        time.Duration // lru的defaultTTL
}

func (c *LRUCache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	/*
		判断c.lru 是否为 nil，如果等于 nil 再创建实例。
		这种方法称之为延迟初始化(Lazy Initialization)，一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。
		主要用于提高性能，并减少程序内存要求。
	.*/
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil, c.ttl)
	}
	c.lru.Add(key, value, c.ttl)
}

func (c *LRUCache) get(key string) (value ByteView, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

type LFUCache struct {
	mu         sync.RWMutex
	lfu        *lfu.LFUCache
	cacheBytes int64
	ttl        time.Duration
}

func (c *LFUCache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lfu == nil {
		c.lfu = lfu.New(c.cacheBytes, nil, c.ttl)
	}
	c.lfu.Add(key, value, c.ttl)
}

func (c *LFUCache) get(key string) (value ByteView, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lfu == nil {
		return
	}
	if v, ok := c.lfu.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
