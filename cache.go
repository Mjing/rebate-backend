package main

import (
	"sync"
	"time"
)

type CacheItem struct {
	RebateProgram RebateProgram
	LoadedAt      time.Time
}

type Cache struct {
	mu          sync.RWMutex
	data        map[uint]CacheItem
	expiration  time.Duration
}

func NewCache(expiration time.Duration) *Cache {
	return &Cache{
		data:       make(map[uint]CacheItem),
		expiration: expiration,
	}
}

// Get retrieves a cached rebate program by ID.
func (c *Cache) Get(id uint) (*RebateProgram, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.data[id]
	if !found || time.Since(item.LoadedAt) > c.expiration {
		return nil, false
	}
	return &item.RebateProgram, true
}

// Set adds or updates the cache with a rebate program.
func (c *Cache) Set(id uint, rebateProgram RebateProgram) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[id] = CacheItem{
		RebateProgram: rebateProgram,
		LoadedAt:      time.Now(),
	}
}

// Clear clears the cache for a given rebate program ID.
func (c *Cache) Clear(id uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, id)
}
