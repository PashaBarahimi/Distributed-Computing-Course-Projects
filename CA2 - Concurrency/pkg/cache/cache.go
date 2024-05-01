package cache

import (
	"dist-concurrency/pkg/event"
	"sync"
	"time"
)

type Cache struct {
	cacheItems map[string]*CacheItem
	items      *sync.Map
	maxSize    int
	muCache    sync.RWMutex
	muItems    *sync.RWMutex
}

func New(items *sync.Map, muItems *sync.RWMutex, size int) *Cache {
	return &Cache{
		cacheItems: make(map[string]*CacheItem),
		items:      items,
		muItems:    muItems,
		maxSize:    size,
	}
}

func (c *Cache) findLeastRecentlyUsed() string {
	var lru string
	lruTime := time.Now()
	for k, v := range c.cacheItems {
		v.Mu.Lock()
		if v.lastAccess.Before(lruTime) {
			lru = k
			lruTime = v.lastAccess
		}
		v.Mu.Unlock()
	}
	return lru
}

func (c *Cache) evict() {
	lru := c.findLeastRecentlyUsed()
	c.muCache.Lock()
	delete(c.cacheItems, lru)
	c.muCache.Unlock()
}

func (c *Cache) add(item *event.Event) *CacheItem {
	c.muCache.Lock()
	defer c.muCache.Unlock()
	if len(c.cacheItems) >= c.maxSize {
		c.evict()
	}
	ci := &CacheItem{
		Event:      item,
		lastAccess: time.Now(),
	}
	c.cacheItems[item.ID] = ci
	return ci
}

func (c *Cache) get(eventID string) *CacheItem {
	if item, ok := c.cacheItems[eventID]; ok {
		item.Mu.Lock()
		defer item.Mu.Unlock()
		item.lastAccess = time.Now()
		return item
	}
	return nil
}

func (c *Cache) GetEvent(eventID string) *CacheItem {
	if item := c.get(eventID); item != nil {
		return item
	}
	c.muItems.Lock()
	defer c.muItems.Unlock()
	if item := c.get(eventID); item != nil {
		return item
	}
	if item, ok := c.items.Load(eventID); ok {
		if e, ok := item.(*event.Event); ok {
			return c.add(e)
		}
	}
	return nil
}
