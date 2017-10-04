// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"github.com/TheThingsNetwork/api"
	pb "github.com/TheThingsNetwork/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/stats"
	"github.com/rcrowley/go-metrics"
)

type status struct {
	uplink      metrics.Meter
	downlink    metrics.Meter
	activations metrics.Meter
}

func (n *networkServer) InitStatus() {
	n.status = &status{
		uplink:      metrics.NewMeter(),
		downlink:    metrics.NewMeter(),
		activations: metrics.NewMeter(),
	}
}

func (n *networkServer) GetStatus() *pb.Status {
	status := new(pb.Status)
	if n.status == nil {
		return status
	}
	status.System = *stats.GetSystem()
	status.Component = *stats.GetComponent()
	uplink := n.status.uplink.Snapshot()
	status.Uplink = &api.Rates{
		Rate1:  float32(uplink.Rate1()),
		Rate5:  float32(uplink.Rate5()),
		Rate15: float32(uplink.Rate15()),
	}
	downlink := n.status.downlink.Snapshot()
	status.Downlink = &api.Rates{
		Rate1:  float32(downlink.Rate1()),
		Rate5:  float32(downlink.Rate5()),
		Rate15: float32(downlink.Rate15()),
	}
	activations := n.status.activations.Snapshot()
	status.Activations = &api.Rates{
		Rate1:  float32(activations.Rate1()),
		Rate5:  float32(activations.Rate5()),
		Rate15: float32(activations.Rate15()),
	}
	return status
}
