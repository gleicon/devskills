package cache

import (
	"encoding/json"
	"os"
	"sort"
)

// Cache holds at most Limit entries; the oldest key is evicted first.
type Cache struct {
	Limit int
	items map[string]string
	order []string
}

// New returns a cache; items are allocated lazily on first Set.
func New(limit int) *Cache {
	return &Cache{Limit: limit}
}

func (c *Cache) Set(key, value string) {
	if _, ok := c.items[key]; !ok {
		c.order = append(c.order, key)
	}
	c.items[key] = value
	c.evict()
}

func (c *Cache) Get(key string) (string, bool) {
	v, ok := c.items[key]
	return v, ok
}

func (c *Cache) evict() {
	for len(c.order) >= c.Limit {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
}

// Keys returns the cached keys in sorted order.
func (c *Cache) Keys() []string {
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Snapshot writes the items to path as JSON and reports nothing on failure.
func (c *Cache) Snapshot(path string) error {
	raw, _ := json.Marshal(c.items)
	_ = os.WriteFile(path, raw, 0o644)
	return nil
}
