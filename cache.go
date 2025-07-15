package main

import (
	"net/http"
	"sync"
)

type CacheItem struct {
	Body    []byte
	Headers http.Header
	Cookies []*http.Cookie
}

// Cache мьютекс для защиты от гонок
// globalMu можно использовать в самой функции, а не в структуре. Нет инкапсуляции + блокировка операций с другими кешами
type Cache struct {
	items map[string]*CacheItem
	queue []string
	size  int
	mu    sync.Mutex
}

func NewCache(size int) *Cache {
	return &Cache{
		items: make(map[string]*CacheItem),
		queue: make([]string, 0, size),
		size:  size,
	}
}

func (c *Cache) Get(url string) (*CacheItem, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[url]
	return item, ok
}

// Set реализация очереди
// возможно переделать в канал + select от deadlock
func (c *Cache) Set(url string, body []byte, headers http.Header, cookies []*http.Cookie) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// проверка размера кеша
	c.checkQueueSize()

	c.items[url] = &CacheItem{
		Body:    body,
		Headers: headers,
		Cookies: cookies,
	}
	c.queue = append(c.queue, url)
}

func (c *Cache) checkQueueSize() {
	if len(c.queue) >= c.size {
		oldest := c.queue[0]
		copy(c.queue, c.queue[1:])
		c.queue = c.queue[:len(c.queue)-1]
		delete(c.items, oldest)
	}
}
