// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/TheThingsNetwork/api"
	pb "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/rcrowley/go-metrics"
)

type status struct {
	uplink            metrics.Meter
	uplinkUnique      metrics.Meter
	downlink          metrics.Meter
	activations       metrics.Meter
	activationsUnique metrics.Meter
	deduplication     metrics.Histogram
	connectedRouters  metrics.Gauge
	connectedHandlers metrics.Gauge
}

func (b *broker) InitStatus() {
	b.status = &status{
		uplink:            metrics.NewMeter(),
		uplinkUnique:      metrics.NewMeter(),
		downlink:          metrics.NewMeter(),
		activations:       metrics.NewMeter(),
		activationsUnique: metrics.NewMeter(),
		deduplication:     metrics.NewHistogram(metrics.NewUniformSample(512)),
		connectedRouters: metrics.NewFunctionalGauge(func() int64 {
			b.routersLock.RLock()
			defer b.routersLock.RUnlock()
			return int64(len(b.routers))
		}),
		connectedHandlers: metrics.NewFunctionalGauge(func() int64 {
			b.handlersLock.RLock()
			defer b.handlersLock.RUnlock()
			return int64(len(b.handlers))
		}),
	}
}

func (b *broker) GetStatus() *pb.Status {
	status := new(pb.Status)
	if b.status == nil {
		return status
	}
	status.System = stats.GetSystem()
	status.Component = stats.GetComponent()
	uplink := b.status.uplink.Snapshot()
	status.Uplink = &api.Rates{
		Rate1:  float32(uplink.Rate1()),
		Rate5:  float32(uplink.Rate5()),
		Rate15: float32(uplink.Rate15()),
	}
	uplinkUnique := b.status.uplinkUnique.Snapshot()
	status.UplinkUnique = &api.Rates{
		Rate1:  float32(uplinkUnique.Rate1()),
		Rate5:  float32(uplinkUnique.Rate5()),
		Rate15: float32(uplinkUnique.Rate15()),
	}
	downlink := b.status.downlink.Snapshot()
	status.Downlink = &api.Rates{
		Rate1:  float32(downlink.Rate1()),
		Rate5:  float32(downlink.Rate5()),
		Rate15: float32(downlink.Rate15()),
	}
	activations := b.status.activations.Snapshot()
	status.Activations = &api.Rates{
		Rate1:  float32(activations.Rate1()),
		Rate5:  float32(activations.Rate5()),
		Rate15: float32(activations.Rate15()),
	}
	activationsUnique := b.status.activationsUnique.Snapshot()
	status.UplinkUnique = &api.Rates{
		Rate1:  float32(activationsUnique.Rate1()),
		Rate5:  float32(activationsUnique.Rate5()),
		Rate15: float32(activationsUnique.Rate15()),
	}
	deduplication := b.status.deduplication.Snapshot().Percentiles([]float64{0.01, 0.05, 0.10, 0.25, 0.50, 0.75, 0.90, 0.95, 0.99})
	status.Deduplication = &api.Percentiles{
		Percentile1:  float32(deduplication[0]),
		Percentile5:  float32(deduplication[1]),
		Percentile10: float32(deduplication[2]),
		Percentile25: float32(deduplication[3]),
		Percentile50: float32(deduplication[4]),
		Percentile75: float32(deduplication[5]),
		Percentile90: float32(deduplication[6]),
		Percentile95: float32(deduplication[7]),
		Percentile99: float32(deduplication[8]),
	}
	status.ConnectedRouters = uint32(b.status.connectedRouters.Snapshot().Value())
	status.ConnectedHandlers = uint32(b.status.connectedHandlers.Snapshot().Value())
	return status
}
