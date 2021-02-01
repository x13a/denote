package filecache

import (
	"io/ioutil"
	"sync"
	"time"
)

func New() *cache {
	return &cache{m: make(map[string]*value)}
}

type value struct {
	Content []byte
	Time    time.Time
}

type cache struct {
	mu sync.RWMutex
	m  map[string]*value
}

func (c *cache) Get(path string) *value {
	c.mu.RLock()
	value, _ := c.m[path]
	c.mu.RUnlock()
	return value
}

func (c *cache) Set(path string, data []byte) *value {
	value := &value{Content: data, Time: time.Now()}
	c.mu.Lock()
	c.m[path] = value
	c.mu.Unlock()
	return value
}

func (c *cache) Has(path string) (ok bool) {
	c.mu.RLock()
	_, ok = c.m[path]
	c.mu.RUnlock()
	return
}

func (c *cache) From(path string) (*value, error) {
	if value := c.Get(path); value != nil {
		return value, nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return c.Set(path, content), nil
}
