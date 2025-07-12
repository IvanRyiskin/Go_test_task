package GoWorkspace

import "sync"

type CacheItem struct {
	Key   string
	Value []byte
}

type Cache struct {
	items map[string][]byte
	queue []string
	size  int
	mu    sync.Mutex
}

func NewCache(size int) *Cache {
	return &Cache{
		items: make(map[string][]byte),
		queue: make([]string, 0, size),
		size:  size,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.items[key]
	return value, ok
}

// реализация очереди
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.queue) >= c.size {
		oldest := c.queue[0]
		copy(c.queue, c.queue[1:])
		c.queue = c.queue[:len(c.queue)-1]
		delete(c.items, oldest)
	}

	c.items[key] = value
	c.queue = append(c.queue, key)
}
