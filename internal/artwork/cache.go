package artwork

import "sync"

// entry is a node in the LRU's intrusive doubly-linked list. Using a typed list
// (rather than container/list's any-typed elements) avoids type assertions.
type entry struct {
	key  string
	url  string
	prev *entry
	next *entry
}

// lruCache is a small thread-safe LRU mapping a lookup key to a resolved public
// artwork URL. An empty string is a valid, cached "no artwork found" result.
// head is the most-recently-used node, tail the least.
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

// get returns the cached URL and whether the key was present, promoting a hit
// to most-recently-used.
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

// put stores url for key, evicting the least-recently-used entry if over capacity.
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
