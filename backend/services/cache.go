package services

import (
	"sync"
	"time"
)

type ttlCache[T any] struct {
	mu     sync.RWMutex
	ttl    time.Duration
	now    func() time.Time
	values map[string]ttlCacheEntry[T]
}

type ttlCacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

func newTTLCache[T any](ttl time.Duration) *ttlCache[T] {
	return &ttlCache[T]{
		ttl:    ttl,
		now:    time.Now,
		values: make(map[string]ttlCacheEntry[T]),
	}
}

func (c *ttlCache[T]) get(key string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}

	now := c.now()
	c.mu.RLock()
	entry, ok := c.values[key]
	c.mu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		return zero, false
	}
	return entry.value, true
}

func (c *ttlCache[T]) set(key string, value T) {
	if c == nil {
		return
	}

	c.mu.Lock()
	c.values[key] = ttlCacheEntry[T]{
		value:     value,
		expiresAt: c.now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *ttlCache[T]) clear() {
	if c == nil {
		return
	}

	c.mu.Lock()
	c.values = make(map[string]ttlCacheEntry[T])
	c.mu.Unlock()
}
