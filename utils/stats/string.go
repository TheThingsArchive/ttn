package stats

import (
	"sync"

	"github.com/rcrowley/go-metrics"
)

// String stores a string
type String interface {
	Get() map[string]string
	Set(string, string)
	Snapshot() String
}

// GetOrRegisterString returns an existing String or constructs and registers a
// new StandardString.
func GetOrRegisterString(name string, r metrics.Registry) String {
	if nil == r {
		r = metrics.DefaultRegistry
	}
	return r.GetOrRegister(name, NewString).(String)
}

// NewString constructs a new StandardString.
func NewString() String {
	if metrics.UseNilMetrics {
		return NilString{}
	}
	s := newStandardString()
	return s
}

// NewRegisteredString constructs and registers a new StandardString.
func NewRegisteredString(name string, r metrics.Registry) String {
	c := NewString()
	if nil == r {
		r = metrics.DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// StringSnapshot is a read-only copy of another String.
type StringSnapshot struct {
	values map[string]string
}

// Get returns the value of events at the time the snapshot was taken.
func (m *StringSnapshot) Get() map[string]string { return m.values }

// Set panics.
func (*StringSnapshot) Set(t, v string) {
	panic("Set called on a StringSnapshot")
}

// Snapshot returns the snapshot.
func (m *StringSnapshot) Snapshot() String { return m }

// NilString is a no-op String.
type NilString struct{}

// Get is a no-op.
func (NilString) Get() map[string]string { return map[string]string{} }

// Set is a no-op.
func (NilString) Set(t, s string) {}

// Snapshot is a no-op.
func (NilString) Snapshot() String { return NilString{} }

// StandardString is the standard implementation of a String.
type StandardString struct {
	lock     sync.RWMutex
	snapshot *StringSnapshot
}

func newStandardString() *StandardString {
	return &StandardString{
		snapshot: &StringSnapshot{
			values: map[string]string{},
		},
	}
}

// Get returns the number of events recorded.
func (m *StandardString) Get() map[string]string {
	m.lock.RLock()
	value := m.snapshot.values
	m.lock.RUnlock()
	return value
}

// Set sets the String to the given value.
func (m *StandardString) Set(tag, str string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.snapshot.values[tag] = str
}

// Snapshot returns a read-only copy of the string.
func (m *StandardString) Snapshot() String {
	m.lock.RLock()
	snapshot := *m.snapshot
	m.lock.RUnlock()
	return &snapshot
}
