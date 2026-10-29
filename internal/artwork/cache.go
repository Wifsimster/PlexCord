package artwork

import "sync"

type entry struct {
	key  string
	url  string
	prev *entry
	next *entry
}

type lruCache struct {
	mu    sync.Mutex
	max   int
	head  *entry
	tail  *entry
	items map[string]*entry
}

func newLRUCache(max int) *lruCache {
	if max <= 0 {
		max = 256
	}
	return &lruCache{max: max, items: make(map[string]*entry, max)}
}

func (c *lruCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok {
		return "", false
	}
	c.moveToFront(e)
	return e.url, true
}

func (c *lruCache) put(key, url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[key]; ok {
		e.url = url
		c.moveToFront(e)
		return
	}
	e := &entry{key: key, url: url}
	c.items[key] = e
	c.pushFront(e)
	if len(c.items) > c.max {
		c.evictLocked()
	}
}

func (c *lruCache) pushFront(e *entry) {
	e.prev = nil
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

func (c *lruCache) unlink(e *entry) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
	e.prev, e.next = nil, nil
}

func (c *lruCache) moveToFront(e *entry) {
	if c.head == e {
		return
	}
	c.unlink(e)
	c.pushFront(e)
}

func (c *lruCache) evictLocked() {
	if c.tail == nil {
		return
	}
	lru := c.tail
	c.unlink(lru)
	delete(c.items, lru.key)
}
