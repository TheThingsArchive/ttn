package stats

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/rcrowley/go-metrics"
)

// Ticker has a Tick() func
type Ticker interface {
	Tick()
	SetDefaultTTL(uint)
	SetTTL(string, uint)
	Renew(string)
}

// DefaultTTL for the registry
var DefaultTTL uint = 0

// Registry is the default metrics registry
var Registry = NewRegistry()

// TTNRegistry is a mutex-protected map of names to metrics.
// It is basically an extension of https://github.com/rcrowley/go-metrics/blob/master/registry.go
type TTNRegistry struct {
	metrics    map[string]interface{}
	timeouts   map[string]uint
	defaultTTL uint
	mutex      sync.Mutex
}

// NewRegistry reates a new registry and starts a goroutine for the TTL.
func NewRegistry() metrics.Registry {
	return &TTNRegistry{
		metrics:    make(map[string]interface{}),
		timeouts:   make(map[string]uint),
		defaultTTL: DefaultTTL,
	}
}

// Each calls the given function for each registered metric.
func (r *TTNRegistry) Each(f func(string, interface{})) {
	for name, i := range r.registered() {
		f(name, i)
	}
}

// Get the metric by the given name or nil if none is registered.
func (r *TTNRegistry) Get(name string) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.metrics[name]
}

// GetOrRegister gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
// The interface can be the metric to register if not found in registry,
// or a function returning the metric for lazy instantiation.
func (r *TTNRegistry) GetOrRegister(name string, i interface{}) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if metric, ok := r.metrics[name]; ok {
		return metric
	}
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}
	r.register(name, i)
	return i
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *TTNRegistry) Register(name string, i interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.register(name, i)
}

// RunHealthchecks runs all registered healthchecks.
func (r *TTNRegistry) RunHealthchecks() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, i := range r.metrics {
		if h, ok := i.(metrics.Healthcheck); ok {
			h.Check()
		}
	}
}

// Unregister the metric with the given name.
func (r *TTNRegistry) Unregister(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.metrics, name)
}

// UnregisterAll unregisters all metrics.  (Mostly for testing.)
func (r *TTNRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name := range r.metrics {
		delete(r.metrics, name)
	}
}

// Tick decreases the TTL of all metrics that have one
// If the TTL becomes zero, the metric is deleted
func (r *TTNRegistry) Tick() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name, ttl := range r.timeouts {
		r.timeouts[name] = ttl - 1
		if ttl == 0 {
			delete(r.metrics, name)
			delete(r.timeouts, name)
		}
	}
}

// SetDefaultTTL sets a default TTL of a number of ticks for all new metrics
func (r *TTNRegistry) SetDefaultTTL(ticks uint) {
	r.defaultTTL = ticks
}

// SetTTL sets a TTL of a number of ticks for a given metric
func (r *TTNRegistry) SetTTL(name string, ticks uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.metrics[name]; ok {
		// Ignore if the existing ttl is higher than the new ttl
		if ttl, ok := r.timeouts[name]; ok && ttl > ticks {
			return
		}
		r.timeouts[name] = ticks
	}
}

// Renew sets the TTL of a metric to the default value
func (r *TTNRegistry) Renew(name string) {
	r.SetTTL(name, r.defaultTTL)
}

func (r *TTNRegistry) register(name string, i interface{}) error {
	if _, ok := r.metrics[name]; ok {
		return fmt.Errorf("duplicate metric: %s", name)
	}
	switch i.(type) {
	case metrics.Counter, metrics.Gauge, metrics.GaugeFloat64, metrics.Healthcheck, metrics.Histogram, metrics.Meter, metrics.Timer, String:
		r.metrics[name] = i
		if r.defaultTTL > 0 {
			r.timeouts[name] = r.defaultTTL
		}
	}
	return nil
}

func (r *TTNRegistry) registered() map[string]interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	metrics := make(map[string]interface{}, len(r.metrics))
	for name, i := range r.metrics {
		metrics[name] = i
	}
	return metrics
}
