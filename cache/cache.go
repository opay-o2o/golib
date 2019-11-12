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
	data     interface{}
	createAt int64
	expire   int64
}

type Cache struct {
	sync.RWMutex
	items map[string]*Item
	timer *time.Ticker
}

func NewCache() *Cache {
	c := &Cache{
		items: make(map[string]*Item, initSize),
		timer: time.NewTicker(interval * time.Second),
	}

	go c.run()
	return c
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	if item, ok := c.items[key]; ok && (item.expire == 0 || time.Now().Unix() < item.createAt+item.expire) {
		return item.data, true
	}

	return nil, false
}

func (c *Cache) Set(key string, value interface{}, expire time.Duration) {
	c.Lock()
	defer c.Unlock()

	item := &Item{data: value, createAt: time.Now().Unix()}

	if expire > 0 {
		item.expire = int64(expire / time.Second)
	}

	c.items[key] = item
}

func (c *Cache) Expire(key string, expire time.Duration) {
	c.Lock()
	defer c.Unlock()

	if item, ok := c.items[key]; ok {
		if expire > 0 {
			item.expire = int64(expire / time.Second)
		} else {
			item.expire = 0
		}
	}
}

func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
}

func (c *Cache) cleanup() {
	c.Lock()
	defer c.Unlock()

	nowTime := time.Now().Unix()

	for key, item := range c.items {
		if item.expire > 0 && nowTime > item.createAt+item.expire {
			delete(c.items, key)
		}
	}
}

func (c *Cache) GetAll() map[string]interface{} {
	c.RLock()
	defer c.RUnlock()

	nowTime := time.Now().Unix()
	items := make(map[string]interface{}, len(c.items))

	for key, item := range c.items {
		if item.expire == 0 || nowTime < item.createAt+item.expire {
			items[key] = item.data
		}
	}

	return items
}

func (c *Cache) run() {
	for range c.timer.C {
		c.cleanup()
	}
}
