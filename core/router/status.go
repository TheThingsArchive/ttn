// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"github.com/TheThingsNetwork/api"
	pb "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/rcrowley/go-metrics"
)

type status struct {
	uplink            metrics.Meter
	downlink          metrics.Meter
	activations       metrics.Meter
	gatewayStatus     metrics.Meter
	connectedGateways metrics.Gauge
	connectedBrokers  metrics.Gauge
}

func (r *router) InitStatus() {
	r.status = &status{
		uplink:        metrics.NewMeter(),
		downlink:      metrics.NewMeter(),
		activations:   metrics.NewMeter(),
		gatewayStatus: metrics.NewMeter(),
		connectedGateways: metrics.NewFunctionalGauge(func() int64 {
			r.gatewaysLock.RLock()
			defer r.gatewaysLock.RUnlock()
			return int64(len(r.gateways))
		}),
		connectedBrokers: metrics.NewFunctionalGauge(func() int64 {
			r.brokersLock.RLock()
			defer r.brokersLock.RUnlock()
			return int64(len(r.brokers))
		}),
	}
}

func (r *router) GetStatus() *pb.Status {
	status := new(pb.Status)
	if r.status == nil {
		return status
	}
	status.System = stats.GetSystem()
	status.Component = stats.GetComponent()
	uplink := r.status.uplink.Snapshot()
	status.Uplink = &api.Rates{
		Rate1:  float32(uplink.Rate1()),
		Rate5:  float32(uplink.Rate5()),
		Rate15: float32(uplink.Rate15()),
	}
	downlink := r.status.downlink.Snapshot()
	status.Downlink = &api.Rates{
		Rate1:  float32(downlink.Rate1()),
		Rate5:  float32(downlink.Rate5()),
		Rate15: float32(downlink.Rate15()),
	}
	activations := r.status.activations.Snapshot()
	status.Activations = &api.Rates{
		Rate1:  float32(activations.Rate1()),
		Rate5:  float32(activations.Rate5()),
		Rate15: float32(activations.Rate15()),
	}
	gatewayStatus := r.status.gatewayStatus.Snapshot()
	status.GatewayStatus = &api.Rates{
		Rate1:  float32(gatewayStatus.Rate1()),
		Rate5:  float32(gatewayStatus.Rate5()),
		Rate15: float32(gatewayStatus.Rate15()),
	}
	status.ConnectedGateways = uint32(r.status.connectedGateways.Snapshot().Value())
	status.ConnectedBrokers = uint32(r.status.connectedBrokers.Snapshot().Value())
	return status
}
