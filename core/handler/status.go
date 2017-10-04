// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/api"
	pb "github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/rcrowley/go-metrics"
)

type status struct {
	uplink      metrics.Meter
	downlink    metrics.Meter
	activations metrics.Meter
}

func (h *handler) InitStatus() {
	h.status = &status{
		uplink:      metrics.NewMeter(),
		downlink:    metrics.NewMeter(),
		activations: metrics.NewMeter(),
	}
}

func (h *handler) GetStatus() *pb.Status {
	status := new(pb.Status)
	if h.status == nil {
		return status
	}
	status.System = *stats.GetSystem()
	status.Component = *stats.GetComponent()
	uplink := h.status.uplink.Snapshot()
	status.Uplink = &api.Rates{
		Rate1:  float32(uplink.Rate1()),
		Rate5:  float32(uplink.Rate5()),
		Rate15: float32(uplink.Rate15()),
	}
	downlink := h.status.downlink.Snapshot()
	status.Downlink = &api.Rates{
		Rate1:  float32(downlink.Rate1()),
		Rate5:  float32(downlink.Rate5()),
		Rate15: float32(downlink.Rate15()),
	}
	activations := h.status.activations.Snapshot()
	status.Activations = &api.Rates{
		Rate1:  float32(activations.Rate1()),
		Rate5:  float32(activations.Rate5()),
		Rate15: float32(activations.Rate15()),
	}
	return status
}
