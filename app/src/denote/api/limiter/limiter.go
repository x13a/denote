package limiter

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/x31a/denote/app/src/denote/api/crypto"
	"bitbucket.org/x31a/denote/app/src/denote/api/limiter/ip"
)

func NewIPLimiter(limit int) *IPLimiter {
	salt, err := crypto.RandRead(sha256.Size)
	if err != nil {
		panic(err)
	}
	return &IPLimiter{
		m:     make(map[string]Value),
		salt:  salt,
		limit: limit,
	}
}

type Value struct {
	count int
	time  time.Time
}

type IPLimiter struct {
	mu    sync.RWMutex
	m     map[string]Value
	limit int
	salt  []byte
}

func (l *IPLimiter) Allow(r *http.Request) (v bool) {
	if !l.IsActive() {
		v = true
		return
	}
	ip := ip.FromRequest(r, true)
	if ip == nil {
		return
	}
	hasher := sha256.New()
	hasher.Write(l.salt)
	hasher.Write(ip)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	l.mu.Lock()
	value, _ := l.m[hash]
	if value.count < l.limit {
		v = true
	}
	l.m[hash] = Value{value.count + 1, time.Now()}
	l.mu.Unlock()
	return
}

func (l *IPLimiter) SetLimit(n int) {
	l.mu.Lock()
	l.limit = n
	l.mu.Unlock()
}

func (l *IPLimiter) IsActive() (v bool) {
	l.mu.RLock()
	v = l.limit > 0
	l.mu.RUnlock()
	return
}

func (l *IPLimiter) Cleaner(interval time.Duration, stopChan chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
Loop:
	for {
		select {
		case <-stopChan:
			break Loop
		case <-ticker.C:
			l.Clean(interval)
		}
	}
	stopChan <- struct{}{}
}

func (l *IPLimiter) Clean(interval time.Duration) {
	now := time.Now()
	l.mu.Lock()
	for k, v := range l.m {
		if v.time.Add(interval).Before(now) {
			delete(l.m, k)
		}
	}
	l.mu.Unlock()
}
