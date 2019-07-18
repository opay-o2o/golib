package cache

import (
	"sync"
	"time"
)

const (
	initSize = 4096
	interval = 10
)

type Item struct {
	data   interface{}
	expire *time.Time
}

type Cache struct {
	sync.RWMutex
	items map[interface{}]*Item
	timer *time.Ticker
}

func NewCache() *Cache {
	c := &Cache{
		items: make(map[interface{}]*Item, initSize),
		timer: time.NewTicker(interval * time.Second),
	}

	go c.run()
	return c
}

func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	if item, ok := c.items[key]; ok && (item.expire == nil || time.Now().Before(*item.expire)) {
		return item.data, true
	}

	return nil, false
}

func (c *Cache) Set(key interface{}, value interface{}, expire time.Duration) {
	c.Lock()
	defer c.Unlock()

	item := &Item{data: value}

	if expire > 0 {
		deadline := time.Now().Add(expire)
		item.expire = &deadline
	}

	c.items[key] = item
}

func (c *Cache) Expire(key interface{}, expire time.Duration) {
	c.Lock()
	defer c.Unlock()

	if item, ok := c.items[key]; ok {
		if expire > 0 {
			deadline := time.Now().Add(expire)
			item.expire = &deadline
		} else {
			item.expire = nil
		}
	}
}

func (c *Cache) Delete(key interface{}) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
}

func (c *Cache) cleanup() {
	c.Lock()
	defer c.Unlock()

	nowTime := time.Now()

	for key, item := range c.items {
		if item.expire != nil && nowTime.After(*item.expire) {
			delete(c.items, key)
		}
	}
}

func (c *Cache) run() {
	for range c.timer.C {
		c.cleanup()
	}
}
