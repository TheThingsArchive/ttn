package component

import "sync/atomic"

// Status indicates the health status of this component
type Status int

const (
	// StatusHealthy indicates a healthy component
	StatusHealthy Status = iota
	// StatusUnhealthy indicates an unhealthy component
	StatusUnhealthy
)

// GetStatus gets the health status of the component
func (c *Component) GetStatus() Status {
	return Status(atomic.LoadInt64(&c.status))
}

// SetStatus sets the health status of the component
func (c *Component) SetStatus(status Status) {
	atomic.StoreInt64(&c.status, int64(status))
}
