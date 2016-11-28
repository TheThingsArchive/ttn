package ratelimit

import ratelimit "gopkg.in/bsm/ratelimit.v1"
import "sync"
import "time"

// Registry for rate limiting
type Registry struct {
	rate     int
	per      time.Duration
	mu       sync.RWMutex
	entities map[string]*ratelimit.RateLimiter
}

// NewRegistry returns a new Registry for rate limiting
func NewRegistry(rate int, per time.Duration) *Registry {
	return &Registry{
		entities: make(map[string]*ratelimit.RateLimiter),
		rate:     rate,
		per:      per,
	}
}

func (r *Registry) getOrCreate(id string, createFunc func() *ratelimit.RateLimiter) *ratelimit.RateLimiter {
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

// Limit returns true if the ratelimit for the given entity has been reached
func (r *Registry) Limit(id string) bool {
	return r.getOrCreate(id, func() *ratelimit.RateLimiter {
		return ratelimit.New(r.rate, r.per)
	}).Limit()
}

// LimitWithRate returns true if the ratelimit for the given entity has been reached
// The first time this function is called for this ID, a RateLimiter is created with
// the given settings
func (r *Registry) LimitWithRate(id string, rate int, per time.Duration) bool {
	return r.getOrCreate(id, func() *ratelimit.RateLimiter {
		return ratelimit.New(rate, per)
	}).Limit()
}
