// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"time"

	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

func (g *Gateway) HandleUplink(uplink *pb_router.UplinkMessage) (err error) {
	toa, err := toa.Compute(uplink)
	if err != nil {
		return err
	}
	g.uplink.Add(time.Now(), uint64(toa))

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.frequencyPlan == nil {
		err = g.setFrequencyPlan(band.Guess(uplink.GetGatewayMetadata().GetFrequency()))
		if err != nil {
			uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, fmt.Sprintf("Could not guess gateway frequency plan: %s", err))
		}
	}
	if g.schedule != nil {
		g.schedule.Sync(uplink.GatewayMetadata.Timestamp)
	}
	g.lastSeen = time.Now()

	// Inject Gateway Metadata
	uplink.GatewayMetadata.GatewayId = g.ID
	uplink.GatewayMetadata.GatewayTrusted = g.status.GatewayTrusted
	if uplink.GatewayMetadata.Gps == nil {
		uplink.GatewayMetadata.Gps = g.status.Gps
	}

	// Inject lorawan metadata
	if lorawan := uplink.GetProtocolMetadata().GetLorawan(); lorawan != nil && g.frequencyPlan != nil {
		lorawan.FrequencyPlan = g.frequencyPlan.Plan
	}

	return nil
}
