// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
)

func (r *router) HandleGatewayStatus(gatewayID string, status *pb_gateway.Status) (err error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID)
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle gateway status")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled gateway status")
		}
	}()
	r.status.gatewayStatus.Mark(1)
	status.Router = r.Identity.Id
	return r.getGateway(gatewayID).HandleStatus(status)
}
