package stats

import "github.com/rcrowley/go-metrics"

// MarkMeter registers an event
func MarkMeter(name string) {
	metrics.GetOrRegisterMeter(name, metrics.DefaultRegistry).Mark(1)
}

// UpdateHistogram registers a new value for a histogram
func UpdateHistogram(name string, value int64) {
	metrics.GetOrRegisterHistogram(name, metrics.DefaultRegistry, metrics.NewUniformSample(1000)).Update(value)
}
