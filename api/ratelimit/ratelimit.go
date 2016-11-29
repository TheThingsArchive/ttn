package ratelimit

import (
	"sync"
	"time"

	"github.com/juju/ratelimit"
)

// Registry for rate limiting
type Registry struct {
	rate     int
	per      time.Duration
	mu       sync.RWMutex
	entities map[string]*ratelimit.Bucket
}

// NewRegistry returns a new Registry for rate limiting
func NewRegistry(rate int, per time.Duration) *Registry {
	return &Registry{
		rate:     rate,
		per:      per,
		entities: make(map[string]*ratelimit.Bucket),
	}
}

func (r *Registry) getOrCreate(id string, createFunc func() *ratelimit.Bucket) *ratelimit.Bucket {
	r.mu.RLock()
	limiter, ok := r.entities[id]
	r.mu.RUnlock()
	if ok {
		return limiter
	}
	limiter = createFunc()
	r.mu.Lock()
	r.entities[id] = limiter
	r.mu.Unlock()
	return limiter
}

func (r *Registry) newFunc() *ratelimit.Bucket {
	return ratelimit.NewBucket(r.per, int64(r.rate))
}

// Limit returns true if the ratelimit for the given entity has been reached
func (r *Registry) Limit(id string) bool {
	return r.Wait(id) != 0
}

// Wait returns the time to wait until available
func (r *Registry) Wait(id string) time.Duration {
	return r.getOrCreate(id, r.newFunc).Take(1)
}
