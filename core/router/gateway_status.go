// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
)

func (r *router) HandleGatewayStatus(gatewayID string, status *pb_gateway.Status) (err error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID)
	start := time.Now()
	var gateway *gateway.Gateway
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle gateway status")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled gateway status")
		}
		if gateway != nil && gateway.MonitorStream != nil {
			gateway.MonitorStream.Send(status)
		}
	}()
	r.status.gatewayStatus.Mark(1)
	status.Router = r.Identity.ID
	gateway = r.getGateway(gatewayID)
	return gateway.HandleStatus(status)
}
