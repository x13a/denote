package cache

import (
	"io/ioutil"
	"sync"
	"time"
)

func NewFileCache() *FileCache {
	return &FileCache{m: make(map[string]*Value)}
}

type Value struct {
	Content []byte
	Time    time.Time
}

type FileCache struct {
	mu sync.RWMutex
	m  map[string]*Value
}

func (c *FileCache) Get(path string) *Value {
	c.mu.RLock()
	value, _ := c.m[path]
	c.mu.RUnlock()
	return value
}

func (c *FileCache) Set(path string, data []byte) *Value {
	value := &Value{Content: data, Time: time.Now()}
	c.mu.Lock()
	c.m[path] = value
	c.mu.Unlock()
	return value
}

func (c *FileCache) Has(path string) (ok bool) {
	c.mu.RLock()
	_, ok = c.m[path]
	c.mu.RUnlock()
	return
}

func (c *FileCache) From(path string) (*Value, error) {
	if value := c.Get(path); value != nil {
		return value, nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return c.Set(path, content), nil
}
